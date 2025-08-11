package sitemaps

import (
	"factual-docs/internal/config"
	postsRepo "factual-docs/internal/repositories/posts"
	"factual-docs/internal/shared/redis"
	"factual-docs/internal/ui"
)

type Service struct {
	postsRepo *postsRepo.Repository
	rdb       redis.Service
	ui        ui.Service
	config    *config.Config
}

func New(
	postsRepo *postsRepo.Repository,
	rdb redis.Service,
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
