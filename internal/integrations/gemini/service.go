// Package gemini creates a gemini service Service
package gemini

import (
	"context"
	"errors"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"

	"google.golang.org/genai"
)

type model struct {
	name    string
	limiter *GeminiLimiter
}

// Service is Gemini struct
type Service struct {
	client   *genai.Client
	models   []model
	catNames []string
}

// New creates new Gemini service
func New(
	ctx context.Context,
	cfg *config.Config,
	rs *rdb.Service,
	catNames []string) (*Service, error) {

	client, err := genai.NewClient(
		ctx, &genai.ClientConfig{
			APIKey: cfg.GeminiAPIKey,
			// Use "v1beta" until "v1" enables agentic media processing.
			HTTPOptions: genai.HTTPOptions{APIVersion: "v1beta"},
		})

	if err != nil {
		return nil, err
	}

	models := make([]model, 0, len(cfg.GeminiModels))
	for _, name := range cfg.GeminiModels {
		if quota, ok := allowedModels[name]; ok {
			limiter, err := NewLimiter(rs, name, cfg.GeminiTimezone, &quota)
			if err != nil {
				return nil, err
			}
			models = append(models, model{name, limiter})
		}
	}

	if len(models) == 0 {
		return nil, errors.New("no valid model(s) supplied")
	}

	return &Service{client, models, catNames}, nil
}
