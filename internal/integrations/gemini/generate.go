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

	response.Intro.OriginalTitle = utils.NormalizeTitle(response.Intro.OriginalTitle, utils.VideoTitleCutoffs)
	response.Main.Summary = utils.NormalizeDescription(response.Main.Summary)

	var directors []string
	for _, director := range response.Intro.Directors {

		name, err := utils.NormalizeName(director)
		if err == nil {
			directors = append(directors, name)
			continue
		}
		slog.ErrorContext(
			ctx,
			"failed to normalize director's name",
			"original_director_name", director,
			"original_title", response.Intro.OriginalTitle,
			"error", err,
		)
	}

	for _, director := range response.Outro.Directors {

		name, err := utils.NormalizeName(director)
		if err != nil {
			slog.ErrorContext(
				ctx,
				"failed to normalize director's name",
				"original_director_name", director,
				"original_title", response.Intro.OriginalTitle,
				"error", err,
			)
			continue
		}

		if !slices.Contains(directors, director) {
			directors = append(directors, name)
		}
	}

	response.Intro.Directors = directors

	return &response, nil
}

// GeneratePostSummary generates post summary and category
func (s *Service) GeneratePostContent(
	ctx context.Context,
	post *models.Post,
	retryConfig *utils.RetryConfig) error {

	mainContents, err := s.MakeVideoContents(
		post.VideoID, models.VideoPartConfig{},
	)

	if err != nil {
		return fmt.Errorf(
			"failed to create gemini contents on SUMMARY on video %q: %w",
			post.VideoID, err)
	}

	genaiConfig := s.NewGenaiConfig()
	genaiConfig.ResponseSchema = s.Schema()

	genaiResponse, err := s.GenerateContent(
		ctx,
		mainContents,
		genaiConfig,
		retryConfig,
	)

	// TODO: Rework and include the
	// Mark the OCR as done
	// if post.Summary != "" {
	// 		post.Summary += models.OcrFlag
	// }

	if err == nil {
		post.OriginalTitle = genaiResponse.Intro.OriginalTitle
		post.Summary = genaiResponse.Main.Summary
		post.Category = &models.Category{Name: genaiResponse.Main.Category}
		post.Directors = genaiResponse.Intro.Directors
		post.ReleaseYear = genaiResponse.Outro.ReleaseYear
		return nil
	}

	// Check if this is a hard block error by the model
	if _, blocked := errors.AsType[*BlockedError](err); !blocked {
		return fmt.Errorf(
			"failed to generate LLM content on SUMMARY on video %q: %w",
			post.VideoID, err,
		)
	}

	// Make another gemini API call just with a text contents
	slog.ErrorContext(
		ctx,
		"failed to generate LLM content on SUMMARY, trying again with text input",
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

	if err != nil {
		return fmt.Errorf(
			"failed to generate LLM content on SUMMARY on video %q: %w",
			post.VideoID, err,
		)
	}

	post.OriginalTitle = genaiResponse.Intro.OriginalTitle
	post.Summary = genaiResponse.Main.Summary
	post.Category = &models.Category{Name: genaiResponse.Main.Category}
	post.Directors = genaiResponse.Intro.Directors
	post.ReleaseYear = genaiResponse.Outro.ReleaseYear

	return nil
}
