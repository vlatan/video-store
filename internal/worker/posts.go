package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/vlatan/video-store/internal/integrations/gemini"
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
			continue
		}

		// If this is NOT a validation error, stop the process
		if _, ok := errors.AsType[*yt.ValidationError](err); !ok {
			return fmt.Errorf(
				"unexpected error during video %q validation; %w",
				video.Id, err,
			)
		}
	}

	return nil
}

// getValidSourcesVideos gets valid videos for a given playlist ids,
// and stores them in the destMap.
func (w *Worker) getValidSourcesVideos(
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
				"couldn't get items from YouTube for source %q; %w",
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
			if _, ok := errors.AsType[*yt.ValidationError](err); ok {
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

			// If the video is already in ytVideosMap as an orphaned video,
			// or as a duplicate video in one or more playlists,
			// we overwrite it, associate it with one YT playlist.
			// If not we just add new video.
			destMap[video.Id] = w.youtube.NewYouTubePost(video, playlistId)
		}
	}

	return nil
}

// adoptVideos associates database videos with playlists if any.
// Checks against the sourceMap.
// Exits with error only if context ended, any other error is just logged.
func (w *Worker) adoptVideos(
	ctx context.Context,
	videos []*models.Post,
	sourceMap map[string]*models.Post,
) error {

	for _, dbVideo := range videos {

		// Check the context first
		if err := ctx.Err(); err != nil {
			return err
		}

		// Check if DB video exists on YouTube
		ytVideo, exists := sourceMap[dbVideo.VideoID]
		if !exists {
			continue
		}

		// Check if we need to update the video playlist
		if ytVideo.PlaylistID == dbVideo.PlaylistID {
			continue
		}

		rowsAffected, err := w.postsRepo.UpdateSource(
			ctx, dbVideo.VideoID, ytVideo.PlaylistID,
		)
		w.stats.AdoptedDbVideos += rowsAffected

		if err == nil {
			continue
		}

		// Exit early if context ended
		if utils.IsContextErr(err) {
			return err
		}

		slog.ErrorContext(
			ctx,
			"failed to update plaulist on video in DB",
			"videoId", dbVideo.VideoID,
			"error", err,
		)
	}

	return nil
}

// deleteVideos deletes videos from DB which are not in destMap.
// Mutates destMap by deleting valid videos from there.
// Returns a slice of valid videos.
// Exits with error only if context ended, any other error is just logged.
func (w *Worker) deleteVideos(
	ctx context.Context,
	dbVideos []*models.Post,
	destMap map[string]*models.Post,
) ([]*models.Post, error) {

	var validDbVideos []*models.Post
	for _, dbVideo := range dbVideos {

		// Check the context first
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		// If the DB video exists on YouTube keep it as valid
		if _, exists := destMap[dbVideo.VideoID]; exists {
			// Keep valid DB videos
			validDbVideos = append(validDbVideos, dbVideo)

			// Delete valid DB videos from the YT map.
			// In this map ONLY the ones that are not in the DB will remain.
			// Meaning the NEW videos that need to be added.
			delete(destMap, dbVideo.VideoID)

			continue
		}

		// Do not remove any more videos from DB if delete limit was reached
		if len(w.stats.DeletedDbVideos) >= deleteLimit {
			continue
		}

		// Delete the video
		_, err := w.postsRepo.DeletePost(ctx, dbVideo.VideoID)
		if err == nil {
			w.stats.DeletedDbVideos = append(w.stats.DeletedDbVideos, dbVideo.VideoID)
			continue
		}

		// Exit early if context ended
		if utils.IsContextErr(err) {
			return nil, err
		}

		slog.ErrorContext(
			ctx,
			"failed to delete video in DB",
			"videoId", dbVideo.VideoID,
			"error", err,
		)
	}

	return validDbVideos, nil
}

// insertVideos generates data for the videos and inserts them in database
func (w *Worker) insertVideos(ctx context.Context, videos []*models.Post) error {

	// Insert new videos in DB
	for i, video := range videos {

		if i > 0 {
			// Sleep with context in mind for 60-90 seconds.
			// Min sleep needs to be 60s to avoid the genai 250k TPM quota.
			if err := utils.SleepJitter(ctx, 60*time.Second, 90*time.Second); err != nil {
				return err
			}
		}

		// Generate post summary and category
		err := w.gemini.GeneratePostSummary(ctx, video, w.geminiRetryConfig)

		// Exit early only if context ended
		if utils.IsContextErr(err) {
			return err
		}

		// For every other error just log it.
		// Ignore any other errors, we'll insert the video in DB regardless.
		if err != nil {
			slog.ErrorContext(
				ctx,
				"failed to generate LLM post summary/category",
				"videoId", video.VideoID,
				"error", err,
			)
		}

		// If generating summary was succesfull procede to generate OCR
		if err == nil {

			// Sleep with context in mind for 60-90 seconds.
			// Min sleep needs to be 60s to avoid the genai 250k TPM quota.
			if err := utils.SleepJitter(ctx, 60*time.Second, 90*time.Second); err != nil {
				return err
			}

			// Generate post original title, directors and release year
			err := w.gemini.GeneratePostOcr(ctx, video, w.geminiRetryConfig)

			// Exit early only if context ended
			if utils.IsContextErr(err) {
				return err
			}

			// For every other error just log it.
			// Ignore any other errors, we'll insert the video in DB regardless.
			if err != nil {
				slog.ErrorContext(
					ctx,
					"failed to generate LLM post OCR",
					"videoId", video.VideoID,
					"error", err,
				)
			}
		}

		rowsAffected, err := w.postsRepo.InsertPost(ctx, video)
		w.stats.InsertedDbVideos += rowsAffected

		// Exit early if context ended
		if utils.IsContextErr(err) {
			return err
		}

		if err != nil {
			slog.ErrorContext(
				ctx,
				"failed to insert video in DB",
				"videoId", video.VideoID,
				"error", err,
			)
		}
	}

	return nil
}

// updateVideos summarizes videos and updates them in database
func (w *Worker) updateVideos(ctx context.Context, videos []*models.Post) error {

	// Insert new videos in DB
	for _, video := range videos {

		var updated bool

		// If post has no summary or category proceed to generate them
		if video.Summary == "" || video.Category == nil {

			// Sleep with context in mind for 60-90 seconds.
			// Min sleep needs to be 60s to avoid the genai 250k TPM quota.
			if err := utils.SleepJitter(ctx, 60*time.Second, 90*time.Second); err != nil {
				return err
			}

			// Generate post summary and category
			err := w.gemini.GeneratePostSummary(ctx, video, w.geminiRetryConfig)

			// Exit with error only if we need to terminate the worker's job,
			// meaning only if RPD quota reached or context ended.
			if errors.Is(err, gemini.ErrDailyLimitReached) || utils.IsContextErr(err) {
				return fmt.Errorf(
					"failed to generate LLM content on video %q; %w",
					video.VideoID, err,
				)
			}

			// For every other error just log it and move onto the next video
			if err != nil {
				slog.ErrorContext(
					ctx,
					"failed to generate LLM content",
					"videoId", video.VideoID,
					"error", err,
				)
				continue
			}

			updated = true
		}

		// If video has no OCR data proceed to generate
		if !strings.HasSuffix(video.Summary, models.OcrFlag) {

			// Sleep with context in mind for 60-90 seconds.
			// Min sleep needs to be 60s to avoid the genai 250k TPM quota.
			if err := utils.SleepJitter(ctx, 60*time.Second, 90*time.Second); err != nil {
				return err
			}

			// Generate post original title, directors and release year
			err := w.gemini.GeneratePostOcr(ctx, video, w.geminiRetryConfig)

			// Exit with error only if we need to terminate the worker's job,
			// meaning only if RPD quota reached or context ended.
			if errors.Is(err, gemini.ErrDailyLimitReached) || utils.IsContextErr(err) {
				return fmt.Errorf(
					"failed to generate LLM content on video %q; %w",
					video.VideoID, err,
				)
			}

			// For every other error just log it,
			if err != nil {
				slog.ErrorContext(
					ctx,
					"failed to generate LLM post OCR",
					"videoId", video.VideoID,
					"error", err,
				)

				// If the summary was not updated then nothing was updated, move onto the next video
				if !updated {
					continue
				}
			}

			updated = true
		}

		if !updated {
			continue
		}

		rowsAffected, err := w.postsRepo.UpdateGeneratedContent(ctx, video)
		w.stats.UpdatedDbVideos += rowsAffected

		// Exit early if context ended
		if utils.IsContextErr(err) {
			return err
		}

		if err != nil {
			slog.ErrorContext(
				ctx,
				"failed to update generated data in DB",
				"videoId", video.VideoID,
				"error", err,
			)
		}
	}

	return nil
}
