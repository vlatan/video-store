package gemini

import (
	"context"
	"strings"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/repositories/categories"

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

	// Save the slice of categories to this service
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

		// Can't return JSON if using web search
		ResponseMIMEType: "application/json",
		// Tools: []*genai.Tool{{GoogleSearch: &genai.GoogleSearch{}}},

		SafetySettings:    safetySettings,
		ResponseSchema:    s.responseSchema(),
		SystemInstruction: s.systemInstruction(),
		MediaResolution:   genai.MediaResolutionLow,
	}

	return s, nil
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
