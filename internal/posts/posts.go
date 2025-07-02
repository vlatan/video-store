package posts

import (
	"factual-docs/internal/auth"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/database"
	"factual-docs/internal/shared/redis"
	tmpls "factual-docs/internal/shared/templates"
)

type Service struct {
	db     database.Service
	rdb    redis.Service
	tm     tmpls.Service
	config *config.Config
	auth   *auth.Service
}

func New(db database.Service, rdb redis.Service, tm tmpls.Service, config *config.Config, auth *auth.Service) *Service {
	return &Service{
		db:     db,
		rdb:    rdb,
		tm:     tm,
		config: config,
		auth:   auth,
	}
}
