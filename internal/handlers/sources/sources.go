package sources

import (
	"factual-docs/internal/integrations/yt"
	postsRepo "factual-docs/internal/repositories/posts"
	sourcesRepo "factual-docs/internal/repositories/sources"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/redis"
	"factual-docs/internal/shared/view"
)

type Service struct {
	postsRepo   *postsRepo.Repository
	sourcesRepo *sourcesRepo.Repository
	rdb         redis.Service
	view        view.Service
	config      *config.Config
	yt          *yt.Service
}

func New(
	postsRepo *postsRepo.Repository,
	sourcesRepo *sourcesRepo.Repository,
	rdb redis.Service,
	view view.Service,
	config *config.Config,
	yt *yt.Service,
) *Service {
	return &Service{
		postsRepo:   postsRepo,
		sourcesRepo: sourcesRepo,
		rdb:         rdb,
		view:        view,
		config:      config,
		yt:          yt,
	}
}
