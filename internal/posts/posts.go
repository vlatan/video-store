package posts

import (
	"factual-docs/internal/auth"
	"factual-docs/internal/services/config"
	"factual-docs/internal/services/database"
	"factual-docs/internal/services/redis"
	tmpls "factual-docs/internal/services/templates"
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
