package sitemaps

import (
	catsRepo "factual-docs/internal/repositories/categories"
	postsRepo "factual-docs/internal/repositories/posts"
	sourcesRepo "factual-docs/internal/repositories/sources"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/redis"
	"factual-docs/internal/shared/view"
)

type Service struct {
	postsRepo   *postsRepo.Repository
	sourcesRepo *sourcesRepo.Repository
	catsRepo    *catsRepo.Repository
	rdb         redis.Service
	view        view.Service
	config      *config.Config
}

func New(
	postsRepo *postsRepo.Repository,
	sourcesRepo *sourcesRepo.Repository,
	catsRepo *catsRepo.Repository,
	rdb redis.Service,
	view view.Service,
	config *config.Config,
) *Service {
	return &Service{
		postsRepo:   postsRepo,
		sourcesRepo: sourcesRepo,
		catsRepo:    catsRepo,
		rdb:         rdb,
		view:        view,
		config:      config,
	}
}
