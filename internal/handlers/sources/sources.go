package sources

import (
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/redis"
	"github.com/vlatan/video-store/internal/integrations/yt"
	postsRepo "github.com/vlatan/video-store/internal/repositories/posts"
	sourcesRepo "github.com/vlatan/video-store/internal/repositories/sources"
	"github.com/vlatan/video-store/internal/ui"
)

type Service struct {
	postsRepo   *postsRepo.Repository
	sourcesRepo *sourcesRepo.Repository
	rdb         redis.Service
	ui          ui.Service
	config      *config.Config
	yt          *yt.Service
}

func New(
	postsRepo *postsRepo.Repository,
	sourcesRepo *sourcesRepo.Repository,
	rdb redis.Service,
	ui ui.Service,
	config *config.Config,
	yt *yt.Service,
) *Service {
	return &Service{
		postsRepo:   postsRepo,
		sourcesRepo: sourcesRepo,
		rdb:         rdb,
		ui:          ui,
		config:      config,
		yt:          yt,
	}
}
