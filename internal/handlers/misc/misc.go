package misc

import (
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/database"
	"factual-docs/internal/shared/redis"
	"factual-docs/internal/shared/view"
)

type Service struct {
	config *config.Config
	db     database.Service
	rdb    redis.Service
	view   view.Service
}

func New(config *config.Config, db database.Service, rdb redis.Service, view view.Service) *Service {
	return &Service{
		config: config,
		db:     db,
		rdb:    rdb,
		view:   view,
	}
}
