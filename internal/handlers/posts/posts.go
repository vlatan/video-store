package posts

import (
	"factual-docs/internal/config"
	"factual-docs/internal/drivers/redis"
	"factual-docs/internal/integrations/gemini"
	"factual-docs/internal/integrations/yt"
	postsRepo "factual-docs/internal/repositories/posts"
	"factual-docs/internal/ui"
)

type Service struct {
	postsRepo *postsRepo.Repository
	rdb       redis.Service
	ui        ui.Service
	config    *config.Config
	yt        *yt.Service
	gemini    *gemini.Service
}

func New(
	postsRepo *postsRepo.Repository,
	rdb redis.Service,
	ui ui.Service,
	config *config.Config,
	yt *yt.Service,
	gemini *gemini.Service,
) *Service {
	return &Service{
		postsRepo: postsRepo,
		rdb:       rdb,
		ui:        ui,
		config:    config,
		yt:        yt,
		gemini:    gemini,
	}
}
