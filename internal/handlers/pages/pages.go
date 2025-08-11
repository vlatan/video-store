package pages

import (
	"factual-docs/internal/config"
	pagesRepo "factual-docs/internal/repositories/pages"
	"factual-docs/internal/shared/redis"
	"factual-docs/internal/ui"
)

type Service struct {
	pagesRepo *pagesRepo.Repository
	rdb       redis.Service
	ui        ui.Service
	config    *config.Config
}

func New(
	pagesRepo *pagesRepo.Repository,
	rdb redis.Service,
	ui ui.Service,
	config *config.Config,
) *Service {
	return &Service{
		pagesRepo: pagesRepo,
		rdb:       rdb,
		ui:        ui,
		config:    config,
	}
}
