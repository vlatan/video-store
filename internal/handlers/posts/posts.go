package posts

import (
	"factual-docs/internal/handlers/auth"
	"factual-docs/internal/integrations/gemini"
	"factual-docs/internal/integrations/yt"
	postsRepo "factual-docs/internal/repositories/posts"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/redis"
	"factual-docs/internal/shared/view"
)

type Service struct {
	postsRepo *postsRepo.Repository
	rdb       redis.Service
	view      view.Service
	config    *config.Config
	auth      *auth.Service
	yt        *yt.Service
	gemini    *gemini.Service
}

func New(
	postsRepo *postsRepo.Repository,
	rdb redis.Service,
	view view.Service,
	config *config.Config,
	auth *auth.Service,
	yt *yt.Service,
	gemini *gemini.Service,
) *Service {
	return &Service{
		postsRepo: postsRepo,
		rdb:       rdb,
		view:      view,
		config:    config,
		auth:      auth,
		yt:        yt,
		gemini:    gemini,
	}
}
