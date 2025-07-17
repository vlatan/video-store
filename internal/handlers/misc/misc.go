package misc

import (
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/database"
	"factual-docs/internal/shared/redis"
	"factual-docs/internal/shared/ui"
)

type Service struct {
	config *config.Config
	db     database.Service
	rdb    redis.Service
	ui     ui.Service
}

func New(config *config.Config, db database.Service, rdb redis.Service, ui ui.Service) *Service {
	return &Service{
		config: config,
		db:     db,
		rdb:    rdb,
		ui:     ui,
	}
}
