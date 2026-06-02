package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"maps"
	"slices"

	"github.com/vlatan/video-store/internal/integrations/yt"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
	"google.golang.org/api/youtube/v3"
)

// Maximum videos to delete per run
const deleteLimit = 5

// Process processes the videos
func (w *Worker) Process(ctx context.Context) error {

	var valErr *yt.ValidationError

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
	w.stats.FetchedDbSources = len(dbSources)

	// GET THE PLAYLISTS FROM YOUTUBE
	// ###################################################################

	// Extract playlist IDs and create DB sources map
	dbSourcesMap := make(map[string]*models.Source, len(dbSources))
	playlistIDs := make([]string, len(dbSources))
	for i, source := range dbSources {
		dbSourcesMap[source.PlaylistID] = &source
		playlistIDs[i] = source.PlaylistID
	}

	// Fetch playlists from YouTube
	ytSources, err := w.youtube.GetSources(ctx, w.ytRetryConfig, playlistIDs...)
	if err != nil {
		return fmt.Errorf(
			"could not fetch the playlists from YouTube; %w",
			err,
		)
	}
	w.stats.FetchedYtSources = len(ytSources)

	// GET THE CORRESPONDING CHANNELS FROM YOUTUBE
	// ###################################################################

	// Extract channel IDs and create YT sources map
	ytSourcesMap := make(map[string]*youtube.Playlist, len(ytSources))
	channelIDs := make([]string, len(ytSources))
	for i, source := range ytSources {
		ytSourcesMap[source.Id] = source
		channelIDs[i] = source.Snippet.ChannelId
	}

	// Fetch corresponding channels
	channels, err := w.youtube.GetChannels(ctx, w.ytRetryConfig, channelIDs...)
	if err != nil {
		return fmt.Errorf(
			"could not fetch the channels from YouTube; %w",
			err,
		)
	}
	w.stats.FetchedYtChannels = len(channels)

	// Create channels map
	channelsMap := make(map[string]*youtube.Channel, len(channels))
	for _, channel := range channels {
		channelsMap[channel.Id] = channel
	}

	// UPDATE THE PLAYLISTS IN DATABASE
	// ###################################################################

	err = w.updatePlaylists(ctx, ytSourcesMap, channelsMap, dbSourcesMap)

	// This can only be context error
	if err != nil {
		return err
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
	w.stats.FetchedDbVideos = len(dbVideos)

	// Define map that will accumulate all valid YT videos
	ytVideosMap := make(map[string]*models.Post)

	// GET ALL THE ORPHAN VALID VIDEOS FROM YOUTUBE
	// ###################################################################

	// Collect the orphans video IDs
	var orphanVideoIDs []string
	for _, video := range dbVideos {
		if video.PlaylistID == "" {
			orphanVideoIDs = append(orphanVideoIDs, video.VideoID)
		}
	}

	// Get valid orphan videos from YT
	if err = w.getOrphanVideos(ctx, orphanVideoIDs, ytVideosMap); err != nil {
		return err
	}

	// GET ALL THE PLAYLIST VALID VIDEOS FROM YOUTUBE
	// ###################################################################

	// Get valid videos from playlists
	for _, playlistID := range playlistIDs {

		// Check the context first
		if err = ctx.Err(); err != nil {
			return err
		}

		sourceItems, err := w.youtube.GetSourceItems(ctx, w.ytRetryConfig, playlistID)
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
		videosMetadata, err := w.youtube.GetVideos(ctx, w.ytRetryConfig, videoIDs...)
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

			// Validate the video
			err = w.youtube.ValidateYouTubeVideo(video)

			// If this is validation error, skip the video
			if errors.As(err, &valErr) {
				continue
			}

			// If this is any other error, stop the process
			if err != nil {
				return fmt.Errorf(
					"unexpected error during video %q validation; %w",
					video.Id, err,
				)
			}

			// Skip if the video is banned (manually deleted).
			// If error is nil the post is IN the deleted_post database table.
			err = w.postsRepo.IsPostBanned(ctx, video.Id)
			if err == nil {
				continue
			}

			// Exit early if context ended
			if utils.IsContextErr(err) {
				return err
			}

			// If the video is already in ytVideosMap as an orphaned video
			// we overwrite it, associate it with a YT playlist.
			// If not we just add new video.
			ytVideosMap[video.Id] = w.youtube.NewYouTubePost(video, playlistID)
			w.stats.FetchedYtVideos++
		}
	}

	// ASSOCIATE VIDEOS TO PLAYLISTS IN DATABASE
	// ###################################################################

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
			w.stats.AdoptedDbVideos++
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

	// DELETE THE OBSOLETE VIDEOS FROM DATABASE
	// ###################################################################

	// Delete videos in DB
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
		if len(w.stats.DeletedDbVideos) >= deleteLimit {
			continue
		}

		// Delete the video
		_, err = w.postsRepo.DeletePost(ctx, dbVideo.VideoID)
		if err == nil {
			w.stats.DeletedDbVideos = append(w.stats.DeletedDbVideos, dbVideo.VideoID)
			continue
		}

		// Exit early if context ended
		if utils.IsContextErr(err) {
			return err
		}

		log.Printf(
			"Could not delete the video '%s' in DB; %v",
			dbVideo.VideoID, err,
		)
	}

	// INSERT THE NEW VIDEOS IN DATABASE
	// ###################################################################

	// Put new YT videos in a slice.
	// ytVideosMap should now contain only new videos.
	newVideos := slices.Collect(maps.Values(ytVideosMap))

	// Summarize new videos in place
	_, err = w.summarizeVideos(ctx, newVideos)

	// This can only be context error
	if err != nil {
		return err
	}

	// Insert new videos in DB
	for _, newVideo := range newVideos {

		_, err = w.postsRepo.InsertPost(ctx, newVideo)
		if err == nil {
			w.stats.InsertedDbVideos++
			continue
		}

		// Exit early if context ended
		if utils.IsContextErr(err) {
			return err
		}

		log.Printf(
			"Failed to insert video '%s' in DB; %v",
			newVideo.VideoID, err,
		)
	}

	// UPDATE THE EXISTING VIDEOS IN DATABASE
	// ###################################################################

	// Summarize the existing videos in place
	indexes, err := w.summarizeVideos(ctx, validDBVideos)

	// This can only be context error
	if err != nil {
		return err
	}

	// Update the existing DB videos
	for _, index := range indexes {

		video := validDBVideos[index]
		_, err = w.postsRepo.UpdateGeneratedData(ctx, video)
		if err == nil {
			w.stats.UpdatedDbVideos++
			continue
		}

		// Exit early if context ended
		if utils.IsContextErr(err) {
			return err
		}

		log.Printf(
			"Failed to update generated data in DB on video '%s'; %v",
			video.VideoID, err,
		)
	}

	return nil
}
