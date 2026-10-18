// Package gemini creates a gemini service Service
package gemini

import (
	"context"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/domains/taxonomy"
	"github.com/vlatan/video-store/internal/drivers/rdb"

	"google.golang.org/genai"
)

type taxer interface {
	Categories(context.Context) (taxonomy.Categories, error)
}

// Service is Gemini struct
type Service struct {
	config   *config.Config
	client   *genai.Client
	limiter  *GeminiLimiter
	catNames []string
}

// New creates new Gemini service
func New(
	ctx context.Context,
	cfg *config.Config,
	taxer taxer,
	rs *rdb.Service,
) (*Service, error) {
	// Configure new client
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: cfg.GeminiAPIKey})
	if err != nil {
		return nil, err
	}

	// Configure new limiter
	limiter, err := NewLimiter(cfg, rs)
	if err != nil {
		return nil, err
	}

	s := &Service{
		config:  cfg,
		client:  client,
		limiter: limiter,
	}

	// Get the categories from cache or DB
	categories, err := rs.Get(
		ctx,
		"categories",
		s.config.CacheTimeout,
		func() (taxonomy.Categories, error) {
			return taxer.Categories(ctx)
		},
	)

	if err != nil {
		return nil, err
	}

	// Extract the category names
	s.catNames = make([]string, len(categories))
	for i, cat := range categories {
		s.catNames[i] = cat.Name
	}

	return s, nil
}

// ConsumeQuota attempts to consume 1 request from the daily and minute buckets.
// It returns a sentinel error if any of the quotas are full.
func (s *Service) ConsumeQuota(ctx context.Context) error {
	return s.limiter.ConsumeQuota(ctx)
}

// Exhausted returns true if the daily limit has already been hit
func (s *Service) Exhausted(ctx context.Context) bool {
	return s.limiter.Exhausted(ctx)
}
