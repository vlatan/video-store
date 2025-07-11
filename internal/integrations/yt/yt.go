package yt

import (
	"context"
	"factual-docs/internal/shared/config"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type Service struct {
	config  *config.Config
	youtube *youtube.Service
}

// Create new YouTube service
func New(ctx context.Context, config *config.Config) (*Service, error) {
	var co option.ClientOption = option.WithAPIKey(config.YouTubeAPIKey)
	youtube, err := youtube.NewService(ctx, co)

	return &Service{
		config:  config,
		youtube: youtube,
	}, err
}
