package sitemaps

import (
	catsRepo "factual-docs/internal/repositories/categories"
	pagesRepo "factual-docs/internal/repositories/pages"
	postsRepo "factual-docs/internal/repositories/posts"
	sourcesRepo "factual-docs/internal/repositories/sources"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/redis"
	"factual-docs/internal/shared/ui"
)

type Service struct {
	postsRepo   *postsRepo.Repository
	sourcesRepo *sourcesRepo.Repository
	catsRepo    *catsRepo.Repository
	pagesRepo   *pagesRepo.Repository
	rdb         redis.Service
	ui          ui.Service
	config      *config.Config
}

func New(
	postsRepo *postsRepo.Repository,
	sourcesRepo *sourcesRepo.Repository,
	catsRepo *catsRepo.Repository,
	pagesRepo *pagesRepo.Repository,
	rdb redis.Service,
	ui ui.Service,
	config *config.Config,
) *Service {
	return &Service{
		postsRepo:   postsRepo,
		sourcesRepo: sourcesRepo,
		catsRepo:    catsRepo,
		pagesRepo:   pagesRepo,
		rdb:         rdb,
		ui:          ui,
		config:      config,
	}
}
