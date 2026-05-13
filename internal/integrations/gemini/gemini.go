package gemini

import (
	"context"
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
	redisService *rdb.Service,
	catsRepo *categories.Repository) (*Service, error) {

	// Configure new client
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: cfg.GeminiAPIKey})
	if err != nil {
		return nil, err
	}

	// Configure new limiter
	limiter, err := NewLimiter(cfg, redisService)
	if err != nil {
		return nil, err
	}

	s := &Service{
		config:  cfg,
		client:  client,
		limiter: limiter,
	}

	// Get the categories from cache or DB
	categories, err := rdb.GetCachedData(
		ctx,
		redisService,
		"categories",
		s.config.CacheTimeout,
		func() (models.Categories, error) {
			return catsRepo.GetCategories(ctx)
		},
	)

	if err != nil {
		return nil, err
	}

	// Save tle slice of categories to this service
	s.categories = categories

	// Extract the category names
	catNames := make([]string, len(categories))
	for i, cat := range categories {
		catNames[i] = cat.Name
	}

	// Save the categories string to this service
	s.catStr = strings.Join(catNames, ", ")

	// Configure genai
	temp, topP := float32(0.0), float32(0.1)
	s.genaiConfig = &genai.GenerateContentConfig{
		Temperature: &temp,
		TopP:        &topP,

		// Can't return JSON when using web search
		// ResponseMIMEType:  "application/json",
		SafetySettings:    safetySettings,
		ResponseSchema:    s.responseSchema(),
		SystemInstruction: s.systemInstruction(),

		// MediaResolution:  genai.MediaResolutionLow,
		Tools: []*genai.Tool{{GoogleSearch: &genai.GoogleSearch{}}},
	}

	return s, nil
}

// makeContents creates Genai contents
func (s *Service) makeContents(video *models.Post) []*genai.Content {

	// Populate user prompt custom text parts
	var parts []*genai.Part

	if prompt := s.config.GeminiPrompt; prompt != "" {
		prompt = fmt.Sprintf("--- SUMMARY --- \n%s", prompt)
		parts = append(parts, genai.NewPartFromText(prompt))
	}

	// Ready the rest of the parts
	title := sanitizePrompt(video.Title)
	description := sanitizePrompt(video.Description)
	url := "https://www.youtube.com/watch?v=" + video.VideoID

	// Create video data genai prompt parts
	videoParts := []*genai.Part{
		genai.NewPartFromText(fmt.Sprintf("--- TITLE --- \n%s", title)),
		genai.NewPartFromText(fmt.Sprintf("--- DESCRIPTION --- \n%s", description)),
		genai.NewPartFromText(fmt.Sprintf("--- YOUTUBE URL --- \n%s", url)),
		genai.NewPartFromText(fmt.Sprintf(
			"--- CATEGORY --- \nSelect ONE category from these categories: %s.",
			s.catStr,
		)),
	}

	// Append the video prompt parts
	parts = append(parts, videoParts...)

	return []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}
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
	contents := s.makeContents(video)

	// Make the API call
	result, err := utils.Retry(
		ctx, rc, func() (*genai.GenerateContentResponse, error) {
			return s.generateContent(ctx, contents)
		},
	)

	if err != nil {
		return nil, err
	}

	// Parse the text output
	response := parseResponse(result.Text(), s.categories)
	response.Summary += utils.UpdateMarker // REMOVE
	response.Title = video.Title

	return response, nil
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
