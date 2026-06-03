package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/vlatan/video-store/internal/integrations/gemini"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// summarizeVideo summarizes and categorizes a video in place
func (w *Worker) summarizeVideo(
	ctx context.Context,
	video *models.Post) error {

	// UNCOMMENT
	// Nothing to update, summary and category are populated
	// if video.Summary != "" &&
	// 	video.Category != nil &&
	// 	video.Category.Name != "" {
	// 	return nil
	// }

	// REMOVE
	// Nothing to update, summary and category are populated
	if strings.Contains(video.Summary, utils.UpdateMarker) &&
		video.Category != nil &&
		video.Category.Name != "" {
		return nil
	}

	// Check if the worker still owns the lock before an expensive API call
	if err := w.lock.CheckLock(ctx); err != nil {
		return fmt.Errorf(
			"this worker %s does not own the lock anymore; %w",
			w.id, err,
		)
	}

	// Generate content using Gemini
	genaiResponse, err := w.gemini.Summarize(ctx, video, w.geminiRetryConfig)

	// Exit with error if RPD reached or context ended
	if errors.Is(err, gemini.ErrDailyLimitReached) || utils.IsContextErr(err) {
		return fmt.Errorf(
			"gemini content generation on video '%s' failed; %v",
			video.VideoID, err)
	}

	// Sleep for 60-90 seconds.
	// Min sleep needs to be 1m to avoid the genai 250k TPM quota,
	minSleep, maxOffset := 60*time.Second, 30*time.Second
	sleep := minSleep + time.Duration(rand.Intn(int(maxOffset))) // #nosec G404
	if err := utils.SleepContext(ctx, sleep); err != nil {
		return err
	}

	// For every other error we just log and exit
	if err != nil {
		log.Printf(
			"gemini content generation on video '%s' failed; %v",
			video.VideoID, err,
		)
		return nil
	}

	video.OriginalTitle = genaiResponse.OriginalTitle
	video.Summary = genaiResponse.Summary
	video.Category = &models.Category{Name: genaiResponse.Category}

	return nil
}
