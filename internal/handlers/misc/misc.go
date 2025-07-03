package misc

import (
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/database"
	"factual-docs/internal/shared/redis"
	tmpls "factual-docs/internal/shared/templates"
)

type Service struct {
	config *config.Config
	db     database.Service
	rdb    redis.Service
	tm     tmpls.Service
}

func New(config *config.Config, db database.Service, rdb redis.Service, tm tmpls.Service) *Service {
	return &Service{
		config: config,
		db:     db,
		rdb:    rdb,
		tm:     tm,
	}
}
