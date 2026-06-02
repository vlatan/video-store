package worker

import (
	"context"
	"errors"
	"fmt"

	"github.com/vlatan/video-store/internal/integrations/yt"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// getValidVideos gets valid videos from YT for given video ids,
// and stores them in the destMap.
func (w *Worker) getValidVideos(
	ctx context.Context,
	videoIds []string,
	destMap map[string]*models.Post,
) error {

	// Get orphans metadata from YT
	videos, err := w.youtube.GetVideos(ctx, w.ytRetryConfig, videoIds...)
	if err != nil {
		return fmt.Errorf(
			"could not get the orphan videos from YouTube; %w",
			err,
		)
	}

	// Validate the videos
	for _, video := range videos {

		err = w.youtube.ValidateYouTubeVideo(video)

		// If no error this is a valid video
		if err == nil {
			destMap[video.Id] = w.youtube.NewYouTubePost(video, "")
			w.stats.FetchedYtVideos++
			continue
		}

		// If this is NOT a validation error, stop the process
		var valErr *yt.ValidationError
		if !errors.As(err, &valErr) {
			return fmt.Errorf(
				"unexpected error during video %q validation; %w",
				video.Id, err,
			)
		}
	}

	return nil
}

// getValidSourcessVideos gets valid videos for a given playlist ids,
// and stores them in the destMap.
func (w *Worker) getValidSourcessVideos(
	ctx context.Context,
	playlistIds []string,
	destMap map[string]*models.Post,
) error {

	// Get valid videos from playlists
	for _, playlistId := range playlistIds {

		// Check the context first
		if err := ctx.Err(); err != nil {
			return err
		}

		sourceItems, err := w.youtube.GetSourceItems(ctx, w.ytRetryConfig, playlistId)
		if err != nil {
			return fmt.Errorf(
				"couldn't get items from YouTube for source '%s'; %w",
				playlistId, err,
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
				playlistId, err,
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
			var valErr *yt.ValidationError
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
			destMap[video.Id] = w.youtube.NewYouTubePost(video, playlistId)
			w.stats.FetchedYtVideos++
		}
	}

	return nil
}
