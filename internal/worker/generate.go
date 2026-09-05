package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/vlatan/video-store/internal/integrations/gemini"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
	"google.golang.org/genai"
)

// generateContent summarizes and categorizes a video in place.
// In addition to the error it returns a bool flag
// to signify if the video was indeed summarized,
// because the error might be nil even if the video was not summarized.
func (w *Worker) generateContent(
	ctx context.Context,
	video *models.Post) (bool, error) {

	// Nothing to update, summary and category are populated
	// TODO: Need to make this condition different
	// in order to extract the director(s) and production year
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

	// Create the main video contents.
	// The first 40 minutes to keep within the 250k TPM quota.
	// With low resolution and FPS of 1.0.
	// 40x60x1x70 = 168k tokens
	mainContents, err := w.gemini.MakeVideoContents(
		video.VideoID, models.VideoPartConfig{
			EndOffset: min(videoDuration, 40*time.Minute),
		},
	)

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

	genaiConfig := w.gemini.NewGenaiConfig()
	genaiConfig.ResponseSchema = w.gemini.SummarySchema()

	// Generate content using Gemini
	genaiResponse, err := w.gemini.GenerateContent(ctx, mainContents, genaiConfig, w.geminiRetryConfig)

	// Exit with error if RPD reached or context ended
	if errors.Is(err, gemini.ErrDailyLimitReached) || utils.IsContextErr(err) {
		return false, fmt.Errorf(
			"failed to generate content on video %q; %w",
			video.VideoID, err)
	}

	// Check if this is a hard block error by the model
	_, blocked := errors.AsType[*gemini.BlockedError](err)

	// For every other error we just log and exit with nil error.
	// The video was not summarized though.
	if !blocked && err != nil {
		slog.ErrorContext(
			ctx, "failed to generate LLM content",
			"videoId", video.VideoID,
			"error", err,
		)
		return false, nil
	}

	// If blocked make another gemini API call just with a text contents
	if blocked {
		slog.ErrorContext(
			ctx, "failed to generate LLM content, trying again with text contents",
			"videoId", video.VideoID,
			"error", err,
		)

		// Create text contents
		textContents := w.gemini.MakeTextContents(video)

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
		genaiResponse, err = w.gemini.GenerateContent(ctx, textContents, genaiConfig, w.geminiRetryConfig)

		// Exit with error if RPD reached or context ended
		if errors.Is(err, gemini.ErrDailyLimitReached) || utils.IsContextErr(err) {
			return false, fmt.Errorf(
				"failed to generate content on video %q; %w",
				video.VideoID, err)
		}

		video.Summary = genaiResponse.Summary
		video.Category = &models.Category{Name: genaiResponse.Category}

		// Exit beacuse the original request was blocked and we used text contents
		return true, nil
	}

	video.Summary = genaiResponse.Summary
	video.Category = &models.Category{Name: genaiResponse.Category}

	// If not blocked make another two calls to extract other details
	partConfigs := []models.VideoPartConfig{
		{
			// Intro config, the first 5 minutes.
			// Increase the FPS to 2.0 and media resolution level to high.
			// 5x60x2x280 = 168k tokens
			EndOffset:  5 * time.Minute,
			FPS:        new(2.0),
			Resolutuon: genai.PartMediaResolutionLevelMediaResolutionHigh,
		},
		{
			// Outro config, the last 5 minutes.
			// Increase the FPS to 2.0 and media resolution level to high.
			// 5x60x2x280 = 168k tokens
			StartOffset: videoDuration - 5*time.Minute,
			FPS:         new(2.0),
			Resolutuon:  genai.PartMediaResolutionLevelMediaResolutionHigh,
		},
	}

	schemas := []*genai.Schema{
		w.gemini.IntroSchema(),
		w.gemini.OutroSchema(),
	}

	for i, config := range partConfigs {

		// Create video contents but now with just the FIRST or LAST X minutes.
		contents, err := w.gemini.MakeVideoContents(video.VideoID, config)

		// Just log the error and exit cleanly with true, nil.
		if err != nil {
			slog.ErrorContext(
				ctx, "failed to create gemini contents",
				"pass", 1,
				"videoId", video.VideoID,
				"pass", i+2,
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

		// Use appropriate schema
		genaiConfig.ResponseSchema = schemas[i]

		// Generate content using Gemini
		genaiResponse, err = w.gemini.GenerateContent(ctx, contents, genaiConfig, w.geminiRetryConfig)

		// Exit with error if RPD reached or context ended
		if errors.Is(err, gemini.ErrDailyLimitReached) || utils.IsContextErr(err) {
			return false, fmt.Errorf(
				"failed to generate content on video %q; %w",
				video.VideoID, err)
		}

		// For every other error just log it and exit with true, nil
		// because we have a summary generated at this point.
		if err != nil {
			slog.ErrorContext(
				ctx,
				"failed to generate LLM content",
				"pass", i+2,
				"videoId", video.VideoID,
				"error", err,
			)
			return true, nil
		}

		// Append if any additional directors discovered
		for _, director := range genaiResponse.Directors {
			name, err := utils.NormalizeName(director)
			if err != nil {
				slog.ErrorContext(
					ctx,
					"failed to normalize director's name",
					"pass", i+2,
					"videoId", video.VideoID,
					"original_name", director,
					"error", err,
				)
				continue
			}
			if !slices.Contains(video.Directors, name) {
				video.Directors = append(video.Directors, name)
			}
		}

		// Assign original title if any
		if genaiResponse.OriginalTitle != "" {
			video.OriginalTitle = genaiResponse.OriginalTitle
		}

		// Assign release year if any
		if genaiResponse.ReleaseYear != 0 {
			video.ReleaseYear = genaiResponse.ReleaseYear
		}
	}

	return true, nil
}
