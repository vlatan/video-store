package misc

import (
	"factual-docs/internal/config"
	"factual-docs/internal/drivers/database"
	"factual-docs/internal/drivers/redis"
	"factual-docs/internal/ui"
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
