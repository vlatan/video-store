package posts

import (
	"factual-docs/internal/handlers/auth"
	"factual-docs/internal/integrations/yt"
	postsRepo "factual-docs/internal/repositories/posts"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/redis"
	tmpls "factual-docs/internal/shared/templates"
)

type Service struct {
	postsRepo *postsRepo.Repository
	rdb       redis.Service
	tm        tmpls.Service
	config    *config.Config
	auth      *auth.Service
	yt        *yt.Service
}

func New(
	postsRepo *postsRepo.Repository,
	rdb redis.Service,
	tm tmpls.Service,
	config *config.Config,
	auth *auth.Service,
	yt *yt.Service,
) *Service {
	return &Service{
		postsRepo: postsRepo,
		rdb:       rdb,
		tm:        tm,
		config:    config,
		auth:      auth,
		yt:        yt,
	}
}
