package yt

import (
	"context"

	"github.com/vlatan/video-store/internal/config"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type Service struct {
	config  *config.Config
	youtube *youtube.Service
}

// Create new YouTube service
func New(ctx context.Context, config *config.Config) (*Service, error) {

	clientOption := option.WithAPIKey(config.YouTubeAPIKey)
	youtube, err := youtube.NewService(ctx, clientOption)

	if err != nil {
		return nil, err
	}

	return &Service{config, youtube}, nil
}
