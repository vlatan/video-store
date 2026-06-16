package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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
) (*genai.GenerateContentResponse, error) {

	// Consume minute and daily quotas before calling the API
	if err := s.ConsumeQuota(ctx); err != nil {
		return nil, fmt.Errorf("gemini limit reached: %w", err)
	}

	response, err := s.client.Models.GenerateContent(
		ctx,
		s.config.GeminiModel,
		contents,
		s.genaiConfig,
	)

	if err != nil {
		return nil, err
	}

	// Check if there are candidates at all.
	// Gemini can return zero candidates if it applies hard block.
	if len(response.Candidates) == 0 {
		return nil, &BlockedErr{response.PromptFeedback}
	}

	return response, nil
}

// GenerateContent generates content using Gemini.
// Retries number of times depending on the retry config passed.
// Unmarshals the result if any and returns a genai response object.
func (s *Service) GenerateContent(
	ctx context.Context,
	video *models.Post,
	contents []*genai.Content,
	rc *utils.RetryConfig,
) (*models.GenaiResponse, error) {

	// Make the API call
	result, err := utils.Retry(ctx, rc,
		func() (*genai.GenerateContentResponse, error) {
			return s.generateContent(ctx, contents)
		},
		// Exit immediately if no candidates returned or RPD limit reached
		func(err error) bool {
			var target *BlockedErr
			return errors.As(err, &target) || errors.Is(err, ErrDailyLimitReached)
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
