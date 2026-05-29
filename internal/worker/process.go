package worker

import (
	"context"
	"fmt"
	"log"
	"maps"
	"math/rand"
	"slices"
	"strings"
	"time"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
	"google.golang.org/api/youtube/v3"
)

// Process processes the videos
func (w *Worker) Process(ctx context.Context) error {

	// Define retry configs for the external APIs
	ytRetryConfig := &utils.RetryConfig{
		MaxRetries: 3,
		MaxJitter:  time.Second,
		Delay:      time.Second,
	}

	geminiRetryConfig := &utils.RetryConfig{
		MaxRetries: 3,
		MaxJitter:  2 * time.Second,
		Delay:      65 * time.Second,
	}

	// GET ALL THE PLAYLISTS FROM DATABASE
	// ###################################################################

	// Fetch all the playlists from DB
	dbSources, err := w.sourcesRepo.GetSources(ctx)

	if err != nil || len(dbSources) == 0 {
		return fmt.Errorf(
			"could not fetch the sources from DB; rows: %v; %w",
			len(dbSources), err,
		)
	}

	items := utils.Plural(len(dbSources), "playlist")
	log.Printf("Fetched %d %s from DB", len(dbSources), items)

	// GET GIVEN PLAYLISTS FROM YOUTUBE
	// ###################################################################

	// Extract playlist IDs and create DB sources map
	dbSourcesMap := make(map[string]*models.Source, len(dbSources))
	playlistIDs := make([]string, len(dbSources))
	for i, source := range dbSources {
		dbSourcesMap[source.PlaylistID] = &source
		playlistIDs[i] = source.PlaylistID
	}

	// Fetch playlists from YouTube
	ytSources, err := w.youtube.GetSources(ctx, ytRetryConfig, playlistIDs...)
	if err != nil {
		return fmt.Errorf(
			"could not fetch the playlists from YouTube; %w",
			err,
		)
	}

	items = utils.Plural(len(ytSources), "playlist")
	log.Printf("Fetched %d %s from YouTube", len(ytSources), items)

	// GET GIVEN CHANNELS FROM YOUTUBE
	// ###################################################################

	// Extract channel IDs and create YT sources map
	ytSourcesMap := make(map[string]*youtube.Playlist, len(ytSources))
	channelIDs := make([]string, len(ytSources))
	for i, source := range ytSources {
		ytSourcesMap[source.Id] = source
		channelIDs[i] = source.Snippet.ChannelId
	}

	// Fetch corresponding channels
	channels, err := w.youtube.GetChannels(ctx, ytRetryConfig, channelIDs...)
	if err != nil {
		return fmt.Errorf(
			"could not fetch the channels from YouTube; %w",
			err,
		)
	}

	// Create channels map
	channelsMap := make(map[string]*youtube.Channel, len(channels))
	for _, channel := range channels {
		channelsMap[channel.Id] = channel
	}

	// UPDATE THE PLAYLISTS IN DATABASE
	// ###################################################################

	// Update each playlist in DB if change in thumbnails
	var updatedPlaylists int
	for playlistID, ytSource := range ytSourcesMap {

		// Check the context first
		if err = ctx.Err(); err != nil {
			return err
		}

		newSource := w.youtube.NewYouTubeSource(
			ytSource, channelsMap[ytSource.Snippet.ChannelId],
		)

		// Check if channel thumbs or title have changed
		dbChThumbs := dbSourcesMap[playlistID].ChannelThumbnails
		if models.ThumbnailsEqual(dbChThumbs, newSource.ChannelThumbnails) &&
			dbSourcesMap[playlistID].ChannelTitle == newSource.ChannelTitle {
			continue
		}

		_, err = w.sourcesRepo.UpdateSource(ctx, newSource)
		if err == nil {
			updatedPlaylists++
			continue
		}

		// Exit early if context ended
		if utils.IsContextErr(err) {
			return err
		}

		log.Printf(
			"could not update source '%s' in DB; %v",
			newSource.PlaylistID, err,
		)
	}

	if updatedPlaylists > 0 {
		items = utils.Plural(updatedPlaylists, "playlist")
		log.Printf("Updated %d %s", updatedPlaylists, items)
	}

	// GET ALL THE VIDEOS FROM DATABASE
	// ###################################################################

	// Get ALL videos from DB, should be ordered by upload date
	dbVideos, err := w.postsRepo.GetAllPosts(ctx)
	if err != nil || len(dbVideos) == 0 {
		return fmt.Errorf(
			"could not fetch the videos from DB; rows: %v; %w",
			len(dbVideos), err,
		)
	}

	items = utils.Plural(len(dbVideos), "video")
	log.Printf("Fetched %d %s from DB", len(dbVideos), items)

	// GET ALL THE ORPHAN VALID VIDEOS FROM YOUTUBE
	// ###################################################################

	// Collect the orphans video IDs
	var orphanVideoIDs []string
	for _, video := range dbVideos {
		if video.PlaylistID == "" {
			orphanVideoIDs = append(orphanVideoIDs, video.VideoID)
		}
	}

	// Get orphans metadata from YT, start forming valid YT videos map
	ytVideosMap := make(map[string]*models.Post)
	ytOrphanVideos, err := w.youtube.GetVideos(ctx, ytRetryConfig, orphanVideoIDs...)
	if err != nil {
		return fmt.Errorf(
			"could not get the orphan videos from YouTube; %w",
			err,
		)
	}

	// Add valid orphan videos to YT map
	for _, video := range ytOrphanVideos {
		if err = w.youtube.ValidateYouTubeVideo(video); err == nil {
			ytVideosMap[video.Id] = w.youtube.NewYouTubePost(video, "")
		}
	}

	// GET ALL THE PLAYLIST VALID VIDEOS FROM YOUTUBE
	// ###################################################################

	// Get valid videos from playlists
	for _, playlistID := range playlistIDs {

		// Check the context first
		if err = ctx.Err(); err != nil {
			return err
		}

		sourceItems, err := w.youtube.GetSourceItems(ctx, ytRetryConfig, playlistID)
		if err != nil {
			return fmt.Errorf(
				"couldn't get items from YouTube for source '%s'; %w",
				playlistID, err,
			)
		}

		// Collect the video IDs for this source
		var videoIDs []string
		for _, item := range sourceItems {
			videoIDs = append(videoIDs, item.ContentDetails.VideoId)
		}

		// Get all the videos metadata for this source
		videosMetadata, err := w.youtube.GetVideos(ctx, ytRetryConfig, videoIDs...)
		if err != nil {
			return fmt.Errorf(
				"couldn't get videos from YouTube for source %s; %w",
				playlistID, err,
			)
		}

		// Keep only the valid videos
		for _, video := range videosMetadata {

			// Check the context first
			if err = ctx.Err(); err != nil {
				return err
			}

			// Skip if invalid video
			if err = w.youtube.ValidateYouTubeVideo(video); err != nil {
				continue
			}

			// Skip if the video is banned (manually deleted).
			// If error is nil the post is IN the deleted_post table.
			err = w.postsRepo.IsPostBanned(ctx, video.Id)
			if err == nil {
				continue
			}

			// Exit early if context ended
			if utils.IsContextErr(err) {
				return err
			}

			// If the video is already in ytVideosMap as an orphaned video
			// we overwrite it, associate it with a YT playlist
			ytVideosMap[video.Id] = w.youtube.NewYouTubePost(video, playlistID)
		}
	}

	items = utils.Plural(len(ytVideosMap), "video")
	log.Printf("Fetched %d valid %s from YouTube", len(ytVideosMap), items)

	// UPDATE VIDEOS' PLAYLIST IDS IN DATABASE
	// ###################################################################

	var adopted int
	for _, dbVideo := range dbVideos {

		// Check the context first
		if err = ctx.Err(); err != nil {
			return err
		}

		// Check if DB video exists on YouTube
		ytVideo, exists := ytVideosMap[dbVideo.VideoID]
		if !exists {
			continue
		}

		// Check if we need to update the video playlist
		if ytVideo.PlaylistID == dbVideo.PlaylistID {
			continue
		}

		_, err = w.postsRepo.UpdatePlaylist(
			ctx, dbVideo.VideoID, ytVideo.PlaylistID,
		)

		if err == nil {
			adopted++
			continue
		}

		// Exit early if context ended
		if utils.IsContextErr(err) {
			return err
		}

		log.Printf(
			"Failed to update the playlist on video '%s'; %v",
			dbVideo.VideoID, err,
		)
	}

	if adopted > 0 {
		items = utils.Plural(adopted, "video")
		log.Printf("Adopted %d %s", adopted, items)
	}

	// DELETE THE OBSOLETE VIDEOS FROM DATABASE
	// ###################################################################

	// Delete videos in DB
	var deleted int
	var deletedVideoIDs []string
	var validDBVideos []*models.Post
	for _, dbVideo := range dbVideos {

		// Check the context first
		if err = ctx.Err(); err != nil {
			return err
		}

		// If the DB video exists on YouTube keep it as valid
		if _, exists := ytVideosMap[dbVideo.VideoID]; exists {
			// Keep valid DB videos
			validDBVideos = append(validDBVideos, dbVideo)

			// Delete valid DB videos from the YT map.
			// In this map ONLY the ones that are not in the DB will remain.
			// Meaning the NEW videos that need to be added.
			delete(ytVideosMap, dbVideo.VideoID)

			continue
		}

		// Do not remove any more videos from DB if delete limit was reached
		if deleted >= deleteLimit {
			continue
		}

		// Delete the video
		_, err = w.postsRepo.DeletePost(ctx, dbVideo.VideoID)
		if err == nil {
			deletedVideoIDs = append(deletedVideoIDs, dbVideo.VideoID)
			deleted++
			continue
		}

		// Exit early if context ended
		if utils.IsContextErr(err) {
			return err
		}

		log.Printf(
			"could not delete the video '%s' in DB; %v",
			dbVideo.VideoID, err,
		)
	}

	if deleted > 0 {
		items = utils.Plural(deleted, "video")
		log.Printf("Deleted %d %s %v", deleted, items, deletedVideoIDs)

		if deleted >= deleteLimit {
			msg := "WARNING: HIT MAX DELETION LIMIT. "
			msg += "If this persists investigate for bugs."
			log.Println(msg)
		}
	}

	// INSERT THE NEW VIDEOS IN DATABASE
	// ###################################################################

	// Put new YT videos in a slice.
	// ytVideosMap should now contain only new videos.
	newVideos := slices.Collect(maps.Values(ytVideosMap))

	// Summarize new videos in place
	_, err = w.summarizeVideos(ctx, geminiRetryConfig, newVideos)

	if err != nil {
		return err
	}

	// Insert new videos in DB
	var inserted int
	for _, newVideo := range newVideos {

		_, err = w.postsRepo.InsertPost(ctx, newVideo)
		if err == nil {
			inserted++
			continue
		}

		// Exit early if context ended
		if utils.IsContextErr(err) {
			return err
		}

		log.Printf(
			"failed to insert video '%s' in DB; %v",
			newVideo.VideoID, err,
		)
	}

	if inserted > 0 {
		items = utils.Plural(inserted, "video")
		log.Printf("Added %d %s", inserted, items)
	}

	// UPDATE THE EXISTING VIDEOS IN DATABASE
	// ###################################################################

	// Summarize the existing videos in place
	indexes, err := w.summarizeVideos(ctx, geminiRetryConfig, validDBVideos)

	if err != nil {
		return err
	}

	// Update the existing DB videos
	var updated int
	for _, index := range indexes {

		video := validDBVideos[index]
		_, err = w.postsRepo.UpdateGeneratedData(ctx, video)
		if err == nil {
			updated++
			continue
		}

		// Exit early if context ended
		if utils.IsContextErr(err) {
			return err
		}

		log.Printf(
			"failed to update generated data in DB on video '%s'; %v",
			video.VideoID, err,
		)
	}

	if updated > 0 {
		items = utils.Plural(updated, "video")
		log.Printf("Updated %d %s", updated, items)
	}

	return nil
}

// summarizeVideos summarizes and categorizes videos in place,
// and returns ther indicies.
func (w *Worker) summarizeVideos(
	ctx context.Context,
	rc *utils.RetryConfig,
	videos []*models.Post) ([]int, error) {

	var summarizedIndicies []int
	for i, video := range videos {

		// Check the context first
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		// Skip summarizing videos if daily quota was reached
		if w.gemini.Exhausted(ctx) {
			break
		}

		// UNCOMMENT
		// Nothing to update, summary and category are populated
		// if video.Summary != "" &&
		// 	video.Category != nil &&
		// 	video.Category.Name != "" {
		// 	continue
		// }

		// REMOVE
		// Nothing to update, summary and category are populated
		if strings.Contains(video.Summary, utils.UpdateMarker) &&
			video.Category != nil &&
			video.Category.Name != "" {
			continue
		}

		// Sleep for 60-90 seconds.
		// Min sleep needs to be 1m to avoid the genai 250k TPM quota,
		// which at the same time mitigates the 5 RPM quota.
		minSleep, maxOffset := 60*time.Second, 30*time.Second
		sleep := minSleep + time.Duration(rand.Intn(int(maxOffset))) // #nosec G404
		if err := utils.SleepContext(ctx, sleep); err != nil {
			return nil, err
		}

		// Check if we still own the lock before an expensive API call
		if err := w.lock.CheckLock(ctx); err != nil {
			return nil, fmt.Errorf(
				"this worker %s does not own the lock anymore; %w",
				w.id, err,
			)
		}

		// Generate content using Gemini
		genaiResponse, err := w.gemini.Summarize(ctx, video, rc)

		// Exit early if context ended
		if utils.IsContextErr(err) {
			return nil, err
		}

		// Skip the video update if error
		if err != nil {
			log.Printf(
				"gemini content generation on video '%s' failed; %v",
				video.VideoID, err,
			)
			continue
		}

		videos[i].OriginalTitle = genaiResponse.OriginalTitle
		videos[i].Summary = genaiResponse.Summary
		videos[i].Category = &models.Category{Name: genaiResponse.Category}
		summarizedIndicies = append(summarizedIndicies, i)
	}

	return summarizedIndicies, nil
}
