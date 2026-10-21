package youtube

import (
	"context"

	"github.com/vlatan/video-store/internal/config"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type Service struct {
	youtube *youtube.Service
}

// Create new YouTube service
func New(ctx context.Context, cfg *config.Config) (*Service, error) {

	clientOption := option.WithAPIKey(cfg.YouTubeAPIKey)
	youtube, err := youtube.NewService(ctx, clientOption)

	if err != nil {
		return nil, err
	}

	return &Service{youtube}, nil
}
