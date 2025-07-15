package pages

import (
	pagesRepo "factual-docs/internal/repositories/pages"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/redis"
	tmpls "factual-docs/internal/shared/templates"
)

type Service struct {
	pagesRepo *pagesRepo.Repository
	rdb       redis.Service
	tm        tmpls.Service
	config    *config.Config
}

func New(
	pagesRepo *pagesRepo.Repository,
	rdb redis.Service,
	tm tmpls.Service,
	config *config.Config,
) *Service {
	return &Service{
		pagesRepo: pagesRepo,
		rdb:       rdb,
		tm:        tm,
		config:    config,
	}
}
