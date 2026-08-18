package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/vlatan/video-store/internal/integrations/gemini"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// generateContent summarizes and categorizes a video in place.
// In addition to the error it returns a bool flag
// to signify if the video was indeed summarized,
// because the error might be nil even if the video was not summarized.
func (w *Worker) generateContent(
	ctx context.Context,
	video *models.Post) (bool, error) {

	// Nothing to update, summary and category are populated
	if video.Summary != "" &&
		video.Category != nil &&
		video.Category.Name != "" {
		return false, nil
	}

	// Get video duration
	videoDuration, err := video.Duration.Seconds()
	if err != nil || videoDuration == 0 {
		return false, fmt.Errorf(
			"couldn't convert video's %q duration %q to seconds; %w",
			video.VideoID, video.Duration, err,
		)
	}

	// Create video contents
	// The first 40 minutes to keep within the 250k TPM quota
	contents, err := w.gemini.MakeVideoContents(video.VideoID, 0, min(videoDuration, 40*time.Minute))
	if err != nil {
		return false, fmt.Errorf(
			"failed to create gemini contents on video %q; %v",
			video.VideoID, err)
	}

	// Sleep with context in mind for 60-90 seconds.
	// Min sleep needs to be 60s to avoid the genai 250k TPM quota.
	if err := utils.SleepJitter(ctx, 60*time.Second, 90*time.Second); err != nil {
		return false, err
	}

	// Check if the worker still owns the lock before an expensive API call
	if err = w.lock.CheckLock(ctx); err != nil {
		return false, fmt.Errorf(
			"this worker %q does not own the lock anymore; %w",
			w.id, err,
		)
	}

	// Generate content using Gemini
	genaiResponse, err := w.gemini.GenerateContent(ctx, contents, w.geminiRetryConfig)

	// Exit with error if RPD reached or context ended
	if errors.Is(err, gemini.ErrDailyLimitReached) || utils.IsContextErr(err) {
		return false, fmt.Errorf(
			"failed to generate content on video %q; %w",
			video.VideoID, err)
	}

	// Check if this is a hard block error by the model.
	// If so make another gemini API call just with a text contents.
	_, blocked := errors.AsType[*gemini.BlockedError](err)
	if blocked {
		log.Printf(
			"failed to generate content on video %q, "+
				"trying again with text contents: %v",
			video.VideoID, err,
		)

		// Create text contents
		contents = w.gemini.MakeTextContents(video)

		// Sleep with context in mind for 60-90 seconds.
		// Min sleep needs to be 60s to avoid the genai 250k TPM quota.
		if err := utils.SleepJitter(ctx, 60*time.Second, 90*time.Second); err != nil {
			return false, err
		}

		// Check if the worker still owns the lock before an expensive API call
		if err = w.lock.CheckLock(ctx); err != nil {
			return false, fmt.Errorf(
				"this worker %q does not own the lock anymore; %w",
				w.id, err,
			)
		}

		// Generate content using Gemini, but now with text contents
		genaiResponse, err = w.gemini.GenerateContent(ctx, contents, w.geminiRetryConfig)

		// Exit with error if RPD reached or context ended
		if errors.Is(err, gemini.ErrDailyLimitReached) || utils.IsContextErr(err) {
			return false, fmt.Errorf(
				"failed to generate content on video %q; %w",
				video.VideoID, err)
		}
	}

	// For every other error we just log and exit with nil error.
	// The video was not summarized though.
	if err != nil {
		log.Printf(
			"failed to generate content on video %q; %v",
			video.VideoID, err,
		)
		return false, nil
	}

	video.OriginalTitle = genaiResponse.OriginalTitle
	video.Summary = genaiResponse.Summary
	video.Category = &models.Category{Name: genaiResponse.Category}
	video.Credits = genaiResponse.Credits
	video.ReleaseYear = genaiResponse.ReleaseYear

	// If not blocked and the video is more than 40 minutes long,
	// make another call with the ending of the video to extract the credits.
	if !blocked && videoDuration > 40*time.Minute {

		// Create video contents but now with an end offset - the last 10 minutes.
		// Just log error and exit cleanly with true, nil.
		contents, err = w.gemini.MakeVideoContents(video.VideoID, videoDuration-10*time.Minute, videoDuration)
		if err != nil {
			slog.ErrorContext(
				ctx,
				"failed to make video contents",
				"videoId", video.VideoID,
				"error", err,
			)
			return true, nil
		}

		// Sleep with context in mind for 60-90 seconds.
		// Min sleep needs to be 60s to avoid the genai 250k TPM quota.
		if err := utils.SleepJitter(ctx, 60*time.Second, 90*time.Second); err != nil {
			return false, err
		}

		// Check if the worker still owns the lock before an expensive API call
		if err = w.lock.CheckLock(ctx); err != nil {
			return false, fmt.Errorf(
				"this worker %q does not own the lock anymore; %w",
				w.id, err,
			)
		}

		// Generate content using Gemini
		genaiSecondResponse, err := w.gemini.GenerateContent(ctx, contents, w.geminiRetryConfig)

		// Exit with error if RPD reached or context ended
		if errors.Is(err, gemini.ErrDailyLimitReached) || utils.IsContextErr(err) {
			return false, fmt.Errorf(
				"failed to generate content on video %q; %w",
				video.VideoID, err)
		}

		// For every other error just log it
		if err != nil {
			slog.ErrorContext(
				ctx,
				"failed to generate LLM content on the second pass",
				"videoId", video.VideoID,
				"error", err,
			)
		}

		if genaiSecondResponse != nil {
			video.Credits = genaiSecondResponse.Credits
			video.ReleaseYear = genaiSecondResponse.ReleaseYear
		}
	}

	return true, nil
}
