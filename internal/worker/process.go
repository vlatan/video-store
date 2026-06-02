package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"maps"
	"slices"
	"time"

	"github.com/vlatan/video-store/internal/integrations/yt"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
	"google.golang.org/api/youtube/v3"
)

// Maximum videos to delete per run
const deleteLimit = 5

// Process processes the videos
func (w *Worker) Process(ctx context.Context) (WorkerStats, error) {

	var stats WorkerStats
	var valErr *yt.ValidationError

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
		return stats, fmt.Errorf(
			"could not fetch the sources from DB; rows: %v; %w",
			len(dbSources), err,
		)
	}
	stats.FetchedDbSources = len(dbSources)

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
		return stats, fmt.Errorf(
			"could not fetch the playlists from YouTube; %w",
			err,
		)
	}
	stats.FetchedYtSources = len(ytSources)

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
	channels, err := w.youtube.GetChannels(ctx, ytRetryConfig, channelIDs...)
	if err != nil {
		return stats, fmt.Errorf(
			"could not fetch the channels from YouTube; %w",
			err,
		)
	}
	stats.FetchedYtChannels = len(channels)

	// Create channels map
	channelsMap := make(map[string]*youtube.Channel, len(channels))
	for _, channel := range channels {
		channelsMap[channel.Id] = channel
	}

	// UPDATE THE PLAYLISTS IN DATABASE
	// ###################################################################

	stats.UpdatedDbSources, err = w.updatePlaylists(ctx, ytSourcesMap, channelsMap, dbSourcesMap)

	// This can only be context error
	if err != nil {
		return stats, err
	}

	// GET ALL THE VIDEOS FROM DATABASE
	// ###################################################################

	// Get ALL videos from DB, should be ordered by upload date
	dbVideos, err := w.postsRepo.GetAllPosts(ctx)
	if err != nil || len(dbVideos) == 0 {
		return stats, fmt.Errorf(
			"could not fetch the videos from DB; rows: %v; %w",
			len(dbVideos), err,
		)
	}
	stats.FetchedDbVideos = len(dbVideos)

	// GET ALL THE ORPHAN VALID VIDEOS FROM YOUTUBE
	// ###################################################################

	// Collect the orphans video IDs
	var orphanVideoIDs []string
	for _, video := range dbVideos {
		if video.PlaylistID == "" {
			orphanVideoIDs = append(orphanVideoIDs, video.VideoID)
		}
	}

	// Get orphans metadata from YT
	ytOrphanVideos, err := w.youtube.GetVideos(ctx, ytRetryConfig, orphanVideoIDs...)
	if err != nil {
		return stats, fmt.Errorf(
			"could not get the orphan videos from YouTube; %w",
			err,
		)
	}

	// Start filling up the YT videos map with valid videos
	ytVideosMap := make(map[string]*models.Post)
	for _, video := range ytOrphanVideos {

		err = w.youtube.ValidateYouTubeVideo(video)

		// If no error this is a valid video
		if err == nil {
			ytVideosMap[video.Id] = w.youtube.NewYouTubePost(video, "")
			stats.FetchedYtVideos++
			continue
		}

		// If this is NOT a validation error, stop the process
		if !errors.As(err, &valErr) {
			return stats, fmt.Errorf(
				"unexpected error during video %q validation; %w",
				video.Id, err,
			)
		}
	}

	// GET ALL THE PLAYLIST VALID VIDEOS FROM YOUTUBE
	// ###################################################################

	// Get valid videos from playlists
	for _, playlistID := range playlistIDs {

		// Check the context first
		if err = ctx.Err(); err != nil {
			return stats, err
		}

		sourceItems, err := w.youtube.GetSourceItems(ctx, ytRetryConfig, playlistID)
		if err != nil {
			return stats, fmt.Errorf(
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
			return stats, fmt.Errorf(
				"couldn't get videos from YouTube for source %s; %w",
				playlistID, err,
			)
		}

		// Keep only the valid videos
		for _, video := range videosMetadata {

			// Check the context first
			if err = ctx.Err(); err != nil {
				return stats, err
			}

			// Validate the video
			err = w.youtube.ValidateYouTubeVideo(video)

			// If this is validation error, skip the video
			if errors.As(err, &valErr) {
				continue
			}

			// If this is any other error, stop the process
			if err != nil {
				return stats, fmt.Errorf(
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
				return stats, err
			}

			// If the video is already in ytVideosMap as an orphaned video
			// we overwrite it, associate it with a YT playlist.
			// If not we just add new video.
			ytVideosMap[video.Id] = w.youtube.NewYouTubePost(video, playlistID)
			stats.FetchedYtVideos++
		}
	}

	// UPDATE VIDEOS' PLAYLIST IDS IN DATABASE
	// ###################################################################

	for _, dbVideo := range dbVideos {

		// Check the context first
		if err = ctx.Err(); err != nil {
			return stats, err
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
			stats.AdoptedDbVideos++
			continue
		}

		// Exit early if context ended
		if utils.IsContextErr(err) {
			return stats, err
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
			return stats, err
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
		if len(stats.DeletedDbVideos) >= deleteLimit {
			continue
		}

		// Delete the video
		_, err = w.postsRepo.DeletePost(ctx, dbVideo.VideoID)
		if err == nil {
			stats.DeletedDbVideos = append(stats.DeletedDbVideos, dbVideo.VideoID)
			continue
		}

		// Exit early if context ended
		if utils.IsContextErr(err) {
			return stats, err
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
	_, err = w.summarizeVideos(ctx, geminiRetryConfig, newVideos)

	// This can only be context error
	if err != nil {
		return stats, err
	}

	// Insert new videos in DB
	for _, newVideo := range newVideos {

		_, err = w.postsRepo.InsertPost(ctx, newVideo)
		if err == nil {
			stats.InsertedDbVideos++
			continue
		}

		// Exit early if context ended
		if utils.IsContextErr(err) {
			return stats, err
		}

		log.Printf(
			"Failed to insert video '%s' in DB; %v",
			newVideo.VideoID, err,
		)
	}

	// UPDATE THE EXISTING VIDEOS IN DATABASE
	// ###################################################################

	// Summarize the existing videos in place
	indexes, err := w.summarizeVideos(ctx, geminiRetryConfig, validDBVideos)

	// This can only be context error
	if err != nil {
		return stats, err
	}

	// Update the existing DB videos
	for _, index := range indexes {

		video := validDBVideos[index]
		_, err = w.postsRepo.UpdateGeneratedData(ctx, video)
		if err == nil {
			stats.UpdatedDbVideos++
			continue
		}

		// Exit early if context ended
		if utils.IsContextErr(err) {
			return stats, err
		}

		log.Printf(
			"Failed to update generated data in DB on video '%s'; %v",
			video.VideoID, err,
		)
	}

	return stats, nil
}
