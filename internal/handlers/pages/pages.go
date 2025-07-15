package pages

import (
	pagesRepo "factual-docs/internal/repositories/pages"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/redis"
	"factual-docs/internal/shared/view"
)

type Service struct {
	pagesRepo *pagesRepo.Repository
	rdb       redis.Service
	view      view.Service
	config    *config.Config
}

func New(
	pagesRepo *pagesRepo.Repository,
	rdb redis.Service,
	view view.Service,
	config *config.Config,
) *Service {
	return &Service{
		pagesRepo: pagesRepo,
		rdb:       rdb,
		view:      view,
		config:    config,
	}
}
