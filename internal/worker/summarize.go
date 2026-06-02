package worker

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// summarizeVideos summarizes and categorizes videos in place,
// and returns ther indicies.
func (w *Worker) summarizeVideos(
	ctx context.Context,
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
		genaiResponse, err := w.gemini.Summarize(ctx, video, w.geminiRetryConfig)

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
