package sources

import (
	"factual-docs/internal/handlers/auth"
	"factual-docs/internal/integrations/yt"
	sourcesRepo "factual-docs/internal/repositories/sources"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/redis"
	tmpls "factual-docs/internal/shared/templates"
)

type Service struct {
	sourcesRepo *sourcesRepo.Repository
	rdb         redis.Service
	tm          tmpls.Service
	config      *config.Config
	auth        *auth.Service
	yt          *yt.Service
}

func New(
	sourcesRepo *sourcesRepo.Repository,
	rdb redis.Service,
	tm tmpls.Service,
	config *config.Config,
	auth *auth.Service,
	yt *yt.Service,
) *Service {
	return &Service{
		sourcesRepo: sourcesRepo,
		rdb:         rdb,
		tm:          tm,
		config:      config,
		auth:        auth,
		yt:          yt,
	}
}
