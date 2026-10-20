package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/vlatan/video-store/internal/types/duration"
	"github.com/vlatan/video-store/internal/utils/normalize"
	"github.com/vlatan/video-store/internal/utils/retry"
	"github.com/vlatan/video-store/internal/utils/sleep"
	"google.golang.org/genai"
)

// The response from the Genai API
type Response struct {
	Title         string   `json:"video_title"`
	OriginalTitle string   `json:"original_title"`
	Summary       string   `json:"summary"`
	Category      string   `json:"category"`
	Directors     []string `json:"directors"`
	ReleaseYear   int16    `json:"release_year"`
}

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
		return nil, &NoCandidatesError{response.PromptFeedback}
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
	retryConfig *retry.Config,
) (*Response, error) {

	// Make the API call
	result, err := retry.Do(ctx, retryConfig,
		func() (*genai.GenerateContentResponse, error) {
			return s.generateContent(ctx, contents, genaiConfig)
		},
		// Exit immediately if no candidates returned or RPD limit reached
		func(err error) bool {
			_, blocked := errors.AsType[*NoCandidatesError](err)
			return blocked || errors.Is(err, ErrDailyLimitReached)
		},
	)

	if err != nil {
		return nil, err
	}

	var resp Response
	if err = json.Unmarshal([]byte(result.Text()), &resp); err != nil {
		return nil, err
	}

	resp.OriginalTitle = normalize.Title(resp.OriginalTitle, normalize.VideoTitleCutoffs)
	resp.Summary = normalize.Description(resp.Summary)

	var directors []string
	for _, director := range resp.Directors {
		name, err := normalize.Name(director)
		if err == nil {
			directors = append(directors, name)
			continue
		}
		slog.WarnContext(
			ctx,
			"failed to normalize director's name",
			"error", err,
		)
	}
	resp.Directors = directors

	return &resp, nil
}

// GeneratePostSummary generates post summary and category
func (s *Service) GeneratePostSummary(
	ctx context.Context,
	videoID, title, description string,
	retryConfig *retry.Config) (*Response, error) {

	// Create summary contents
	summaryContents := s.NewSummaryContents(videoID)
	genaiConfig := s.NewGenaiConfig()
	genaiConfig.ResponseSchema = s.NewSummarySchema()

	resp, err := s.GenerateContent(
		ctx,
		summaryContents,
		genaiConfig,
		retryConfig,
	)

	if err == nil {
		return resp, nil
	}

	// Check if this is a hard block error by the model
	if _, blocked := errors.AsType[*NoCandidatesError](err); !blocked {
		return nil, fmt.Errorf(
			"failed to generate LLM content on SUMMARY on video %s: %w",
			videoID, err,
		)
	}

	// Make another gemini API call just with a text contents
	slog.ErrorContext(
		ctx,
		"failed to generate LLM content on SUMMARY, trying again with text input",
		"videoId", videoID,
		"error", err,
	)

	// Create text contents
	textContents := s.NewTextContents(videoID, title, description)

	// Sleep with context in mind for 60-90 seconds.
	// Min sleep needs to be 60s to avoid the genai 250k TPM quota.
	if err = sleep.Jitter(ctx, 60*time.Second, 90*time.Second); err != nil {
		return nil, err
	}

	// Generate content using Gemini, but now with text contents
	resp, err = s.GenerateContent(
		ctx,
		textContents,
		genaiConfig,
		retryConfig,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to generate LLM content on SUMMARY on video %s: %w",
			videoID, err,
		)
	}

	return resp, nil
}

// GeneratePostOCR reads original title, directors and release year from screen
func (s *Service) GeneratePostOCR(
	ctx context.Context,
	videoID string,
	duration duration.ISO8601,
	retryConfig *retry.Config) (*Response, error) {

	// Get video duration
	videoDuration, err := duration.Parse()
	if err != nil || videoDuration == 0 {
		return nil, fmt.Errorf(
			"couldn't convert video %s duration %s to seconds; %w",
			videoID, duration, err,
		)
	}

	// Preapare two separate configs
	startOffset, endOffset := videoDuration-200*time.Second, 300*time.Second
	ocrConfigs := []struct {
		desc     string
		schema   *genai.Schema
		contents []*genai.Content
	}{
		{"INTRO", s.NewIntroSchema(), s.NewIntroContents(videoID, endOffset)},
		{"OUTRO", s.NewOutroSchema(), s.NewOutroContents(videoID, startOffset)},
	}

	// Exclude the outro OCR because genai will reject very long start offset
	if startOffset >= 16*time.Hour {
		ocrConfigs = ocrConfigs[:1]
	}

	// Make two calls to extract the OCR details
	var result Response
	genaiConfig := s.NewGenaiConfig()
	for _, cfg := range ocrConfigs {

		// Sleep with context in mind for 60-90 seconds.
		// Min sleep needs to be 60s to avoid the genai 250k TPM quota.
		if err := sleep.Jitter(ctx, 60*time.Second, 90*time.Second); err != nil {
			return nil, err
		}

		// Use appropriate schema in the genai config
		genaiConfig.ResponseSchema = cfg.schema

		// Generate content using Gemini
		resp, err := s.GenerateContent(ctx, cfg.contents, genaiConfig, retryConfig)

		// Move onto the next call if this is a hard block error by the model
		if _, blocked := errors.AsType[*NoCandidatesError](err); blocked {
			slog.ErrorContext(
				ctx,
				fmt.Sprintf("failed to generate LLM content on %s", cfg.desc),
				"videoId", videoID,
				"error", err,
			)
			continue
		}

		if err != nil {
			return nil, fmt.Errorf(
				"failed to generate LLM content on %s on video %s: %w",
				cfg.desc, videoID, err,
			)
		}

		// Assign original title if any, do not overwrite
		if resp.OriginalTitle != "" {
			result.OriginalTitle = resp.OriginalTitle
		}

		// Append if any additional directors discovered
		for _, director := range resp.Directors {
			if !slices.Contains(result.Directors, director) {
				result.Directors = append(result.Directors, director)
			}
		}

		// Assign release year if any, overwrite if found in outro
		maxYear := time.Now().Year() + 1
		if resp.ReleaseYear >= 1900 && int(resp.ReleaseYear) <= maxYear {
			result.ReleaseYear = resp.ReleaseYear
		}

	}

	return &result, nil
}
