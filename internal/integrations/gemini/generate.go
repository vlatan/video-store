package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
	"google.golang.org/genai"
)

// generateContent is a wrapper around the model's GenerateContent method
// which internally handles the daily and minutely rate limit,
// as well as a scenario where no candidates are returned.
func (s *Service) generateContent(
	ctx context.Context,
	contents []*genai.Content,
	genaiConfig *genai.GenerateContentConfig,
) (*genai.GenerateContentResponse, error) {

	// Consume minute and daily quotas before calling the API
	if err := s.ConsumeQuota(ctx); err != nil {
		return nil, fmt.Errorf("gemini limit reached: %w", err)
	}

	response, err := s.client.Models.GenerateContent(
		ctx,
		s.config.GeminiModel,
		contents,
		genaiConfig,
	)

	if err != nil {
		return nil, err
	}

	// Check if there are candidates at all.
	// Gemini can return zero candidates if it applies hard block.
	if len(response.Candidates) == 0 {
		return nil, &BlockedError{response.PromptFeedback}
	}

	return response, nil
}

// GenerateContent generates content using Gemini.
// Retries number of times depending on the retry config passed.
// Unmarshals the result if any and returns a genai response object.
func (s *Service) GenerateContent(
	ctx context.Context,
	contents []*genai.Content,
	genaiConfig *genai.GenerateContentConfig,
	retryConfig *utils.RetryConfig,
) (*models.GenaiResponse, error) {

	// Make the API call
	result, err := utils.Retry(ctx, retryConfig,
		func() (*genai.GenerateContentResponse, error) {
			return s.generateContent(ctx, contents, genaiConfig)
		},
		// Exit immediately if no candidates returned or RPD limit reached
		func(err error) bool {
			_, isBlockedError := errors.AsType[*BlockedError](err)
			return isBlockedError || errors.Is(err, ErrDailyLimitReached)
		},
	)

	if err != nil {
		return nil, err
	}

	var response models.GenaiResponse
	if err = json.Unmarshal([]byte(result.Text()), &response); err != nil {
		return nil, err
	}

	response.Title = utils.NormalizeTitle(response.Title, utils.VideoTitleCutoffs)
	response.OriginalTitle = utils.NormalizeTitle(response.OriginalTitle, utils.VideoTitleCutoffs)
	response.Summary = utils.NormalizeDescription(response.Summary)

	return &response, nil
}

func (s *Service) GeneratePostContent(ctx context.Context, post *models.Post) error {

	videoDuration, err := post.Duration.Seconds()
	if err != nil || videoDuration == 0 {
		return fmt.Errorf("couldn't convert video's duration to seconds: %w", err)
	}

	// Create the main video contents.
	// The first 40 minutes to keep within the 250k TPM quota.
	// With low resolution and FPS of 1.0.
	// 40x60x1x70 = 168k tokens
	mainContents, err := s.MakeVideoContents(
		post.VideoID, models.VideoPartConfig{
			EndOffset: min(videoDuration, 40*time.Minute),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to create gemini contents: %w", err)
	}

	genaiConfig := s.NewGenaiConfig()
	genaiConfig.ResponseSchema = s.SummarySchema()

	retryConfig := &utils.RetryConfig{
		MaxRetries: 1,
		MaxJitter:  2 * time.Second,
		Delay:      65 * time.Second,
	}

	genaiResponse, err := s.GenerateContent(
		ctx,
		mainContents,
		genaiConfig,
		retryConfig,
	)

	// Check if this is a hard block error by the model
	_, blocked := errors.AsType[*BlockedError](err)

	// Exit if fatal error
	if !blocked && err != nil {
		return fmt.Errorf(
			"failed to generate LLM content on video %q: %w", post.VideoID, err,
		)
	}

	// If blocked make another gemini API call just with a text contents
	if blocked {
		slog.ErrorContext(
			ctx, "failed to generate LLM content, trying again with text input",
			"videoId", post.VideoID,
			"error", err,
		)

		// Create text contents
		textContents := s.MakeTextContents(post)

		// Sleep with context in mind for 60-90 seconds.
		// Min sleep needs to be 60s to avoid the genai 250k TPM quota.
		if err := utils.SleepJitter(ctx, 60*time.Second, 90*time.Second); err != nil {
			return err
		}

		// Generate content using Gemini, but now with text contents
		genaiResponse, err = s.GenerateContent(
			ctx,
			textContents,
			genaiConfig,
			retryConfig,
		)

		post.Summary = genaiResponse.Summary
		post.Category = &models.Category{Name: genaiResponse.Category}

		return nil
	}

	post.Summary = genaiResponse.Summary
	post.Category = &models.Category{Name: genaiResponse.Category}

	// Preaparre two separate schemas and part configs
	schemas := []*genai.Schema{s.IntroSchema(), s.OutroSchema()}
	partConfigs := []models.VideoPartConfig{
		{
			// Intro config, the first 5 minutes.
			// Increase the media resolution level to high.
			// 5x60x1x280 = 84k tokens
			EndOffset:  5 * time.Minute,
			Resolutuon: genai.PartMediaResolutionLevelMediaResolutionHigh,
		},
		{
			// Outro config, the last 200 seconds.
			// Increase the FPS to 3.0 and media resolution level to high.
			// 200x3x280 = 168k tokens
			StartOffset: videoDuration - 200*time.Second,
			FPS:         new(3.0),
			Resolutuon:  genai.PartMediaResolutionLevelMediaResolutionHigh,
		},
	}

	// If not blocked make another two calls to extract other details
	for i, config := range partConfigs {

		// Create video contents but now with just the FIRST and LAST x minutes.
		contents, err := s.MakeVideoContents(post.VideoID, config)

		if err != nil {
			slog.ErrorContext(
				ctx, "failed to create gemini contents",
				"videoId", post.VideoID,
				"pass", i+2,
				"error", err,
			)
			// Abort with nil error, no need to continue.
			// Salvage the generated content we already have.
			return nil
		}

		// Sleep with context in mind for 60-90 seconds.
		// Min sleep needs to be 60s to avoid the genai 250k TPM quota.
		if err := utils.SleepJitter(ctx, 60*time.Second, 90*time.Second); err != nil {
			return err
		}

		// Use appropriate schema
		genaiConfig.ResponseSchema = schemas[i]

		// Generate content using Gemini
		genaiResponse, err = s.GenerateContent(ctx, contents, genaiConfig, retryConfig)

		// Exit if context ended
		if utils.IsContextErr(err) {
			return fmt.Errorf("failed to generate LLM content (%d pass): %w", i+2, err)
		}

		// For every other error just log it and exit with nil
		if err != nil {
			slog.ErrorContext(
				ctx,
				"failed to generate LLM content",
				"pass", i+2,
				"videoId", post.VideoID,
				"error", err,
			)
			return nil
		}

		// Append if any additional directors discovered
		for _, director := range genaiResponse.Directors {
			name, err := utils.NormalizeName(director)
			if err != nil {
				slog.ErrorContext(
					ctx,
					"failed to normalize director's name",
					"pass", i+2,
					"videoId", post.VideoID,
					"original_name", director,
					"error", err,
				)
				continue
			}
			if !slices.Contains(post.Directors, name) {
				post.Directors = append(post.Directors, name)
			}
		}

		// Assign original title if any
		if genaiResponse.OriginalTitle != "" {
			post.OriginalTitle = genaiResponse.OriginalTitle
		}

		// Assign release year if any
		if genaiResponse.ReleaseYear != 0 {
			post.ReleaseYear = genaiResponse.ReleaseYear
		}

		slog.InfoContext(
			ctx,
			"video results",
			"pass", i+2,
			"videoId", post.VideoID,
			"original title", genaiResponse.OriginalTitle,
			"directors", genaiResponse.Directors,
			"releaseYear", genaiResponse.ReleaseYear,
		)
	}

	return nil
}
