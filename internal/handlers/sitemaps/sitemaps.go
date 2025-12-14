package sitemaps

import (
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/redis"
	postsRepo "github.com/vlatan/video-store/internal/repositories/posts"
	"github.com/vlatan/video-store/internal/ui"
)

type Service struct {
	postsRepo *postsRepo.Repository
	rdb       *redis.RedisService
	ui        ui.Service
	config    *config.Config
}

func New(
	postsRepo *postsRepo.Repository,
	rdb *redis.RedisService,
	ui ui.Service,
	config *config.Config,
) *Service {
	return &Service{
		postsRepo: postsRepo,
		rdb:       rdb,
		ui:        ui,
		config:    config,
	}
}
