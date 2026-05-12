package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/repositories/categories"
	"github.com/vlatan/video-store/internal/utils"

	"google.golang.org/genai"
)

// Create new Gemini service
func New(
	ctx context.Context,
	cfg *config.Config,
	rdb *rdb.Service,
	catsRepo *categories.Repository) (*Service, error) {

	// Configure new client
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: cfg.GeminiAPIKey})
	if err != nil {
		return nil, err
	}

	limiter, err := NewLimiter(cfg, rdb)
	if err != nil {
		return nil, err
	}

	s := &Service{
		config:   cfg,
		client:   client,
		limiter:  limiter,
		rdb:      rdb,
		catsRepo: catsRepo,
	}

	temp, topP := float32(0.0), float32(0.1)
	s.genaiConfig = &genai.GenerateContentConfig{
		Temperature: &temp,
		TopP:        &topP,

		// Can't return JSON when using web search
		// ResponseMIMEType:  "application/json",
		SafetySettings:    safetySettings,
		ResponseSchema:    s.responseSchema(ctx),
		SystemInstruction: s.systemInstruction(),

		// MediaResolution:  genai.MediaResolutionLow,
		Tools: []*genai.Tool{{GoogleSearch: &genai.GoogleSearch{}}},
	}

	return s, nil
}

// systemInstruction generates system instructions
func (s *Service) systemInstruction() *genai.Content {
	content := []string{
		"Write as if you are a historian or journalist reporting on the subject matter itself.",
		"Write in third-person factual prose, as if writing for a news article.",
		"Never use hedging language. Use specific, verifiable facts only.",
		"If a fact cannot be stated with confidence, omit it entirely.",
		"Do not use transitional or connective filler between facts.",
		"State each fact as a direct sentence.",
		"Do NOT mention the given media itself - write about its SUBJECT.",
		"Avoid: flowery language, metaphors, purple prose, and generalized statements.",
		"Do not include timestamps.",
		"Do not use UPPER CASE.",
		"Do not use em dashes (—).",
	}

	contentText := strings.Join(content, "\n")
	return genai.NewContentFromText(contentText, genai.RoleUser)
}

// Generate Genai content
func (s *Service) generateContent(
	ctx context.Context,
	contents []*genai.Content,
) (*genai.GenerateContentResponse, error) {

	// Consume minute and daily quotas before calling the API
	if err := s.AcquireQuota(ctx); err != nil {
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

// Create the prompt and generate content using Gemini
// https://ai.google.dev/gemini-api/docs/video-understanding#youtube
func (s *Service) Summarize(
	ctx context.Context,
	video *models.Post,
	rc *utils.RetryConfig,
) (*models.GenaiResponse, error) {

	// Make Genai contents
	contents, err := s.makeContents(ctx, video)
	if err != nil {
		return nil, fmt.Errorf("failed to make Genai contents; %w", err)
	}

	// Make the API call
	result, err := utils.Retry(
		ctx, rc, func() (*genai.GenerateContentResponse, error) {
			return s.generateContent(ctx, contents)
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to generate Genai content; %w", err)
	}

	var response models.GenaiResponse
	if err := json.Unmarshal([]byte(result.Text()), &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Genai JSON response; %w", err)
	}

	// Add marker to summary
	response.Summary += utils.UpdateMarker // REMOVE
	return &response, nil
}

// makeContents creates Genai contents
func (s *Service) makeContents(
	ctx context.Context,
	video *models.Post,
) ([]*genai.Content, error) {

	// Gather the media parts
	var parts []*genai.Part
	contents := []*genai.Content{genai.NewContentFromParts(parts, genai.RoleUser)}
	return contents, nil
}

// AcquireQuota attempts to consume 1 request from the daily and minute buckets.
// It returns a sentinel error if any of the quotas are full.
func (s *Service) AcquireQuota(ctx context.Context) error {
	return s.limiter.AcquireQuota(ctx)
}

// Exhausted returns true if the daily limit has already been hit
func (s *Service) Exhausted(ctx context.Context) bool {
	return s.limiter.Exhausted(ctx)
}
