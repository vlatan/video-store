// Package gemini creates a gemini service Service
package gemini

import (
	"context"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"

	"google.golang.org/genai"
)

// Service is Gemini struct
type Service struct {
	cfg      *config.Config
	client   *genai.Client
	limiter  *GeminiLimiter
	catNames []string
}

// New creates new Gemini service
func New(
	ctx context.Context,
	cfg *config.Config,
	rs *rdb.Service,
	catNames []string,
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
		cfg:      cfg,
		client:   client,
		limiter:  limiter,
		catNames: catNames,
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
