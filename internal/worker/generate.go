package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
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

	// Create video contents
	contents, err := w.gemini.MakeVideoContents(video)
	if err != nil {
		return false, fmt.Errorf(
			"failed to create gemini contents on video %q; %v",
			video.VideoID, err)
	}

	// Check if the worker still owns the lock before an expensive API call
	if err = w.lock.CheckLock(ctx); err != nil {
		return false, fmt.Errorf(
			"this worker %q does not own the lock anymore; %w",
			w.id, err,
		)
	}

	// Generate content using Gemini
	genaiResponse, err := w.gemini.GenerateContent(ctx, video, contents, w.geminiRetryConfig)

	// Exit with error if RPD reached or context ended
	if errors.Is(err, gemini.ErrDailyLimitReached) || utils.IsContextErr(err) {
		return false, fmt.Errorf(
			"failed to generate content on video %q; %w",
			video.VideoID, err)
	}

	// Sleep with context in mind for 60-90 seconds.
	// Min sleep needs to be 60s to avoid the genai 250k TPM quota.
	if err := sleep(ctx, 60*time.Second, 90*time.Second); err != nil {
		return false, err
	}

	// Check if this is a hard block error by the model.
	// If so make another gemini API call just with a text contents.
	var target *gemini.BlockedError
	if errors.As(err, &target) {
		log.Printf(
			"failed to generate content on video %q, "+
				"trying again with text contents: %v",
			video.VideoID, err,
		)

		// Create text contents
		contents = w.gemini.MakeTextContents(video)

		// Check if the worker still owns the lock before an expensive API call
		if err = w.lock.CheckLock(ctx); err != nil {
			return false, fmt.Errorf(
				"this worker %q does not own the lock anymore; %w",
				w.id, err,
			)
		}

		// Generate content using Gemini, but now with text contents
		genaiResponse, err = w.gemini.GenerateContent(ctx, video, contents, w.geminiRetryConfig)

		// Exit with error if RPD reached or context ended
		if errors.Is(err, gemini.ErrDailyLimitReached) || utils.IsContextErr(err) {
			return false, fmt.Errorf(
				"failed to generate content on video %q; %w",
				video.VideoID, err)
		}

		// Sleep with context in mind for 60-90 seconds.
		// Min sleep needs to be 60s to avoid the genai 250k TPM quota.
		if err := sleep(ctx, 60*time.Second, 90*time.Second); err != nil {
			return false, err
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

	return true, nil
}

// sleep sleeps with context in mind
// for a random duration between min and max sleep time
func sleep(ctx context.Context, minSleep, maxSleep time.Duration) error {
	if maxSleep < minSleep {
		return errors.New("max sleep time < min sleep time")
	}

	if maxSleep == minSleep {
		return utils.SleepContext(ctx, minSleep)
	}

	sleepTime := minSleep + rand.N(maxSleep-minSleep) // #nosec G404
	return utils.SleepContext(ctx, sleepTime)
}
