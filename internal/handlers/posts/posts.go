package posts

import (
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/integrations/gemini"
	"github.com/vlatan/video-store/internal/integrations/yt"
	postsRepo "github.com/vlatan/video-store/internal/repositories/posts"
	"github.com/vlatan/video-store/internal/ui"
)

type Service struct {
	postsRepo *postsRepo.Repository
	rdb       *rdb.Service
	ui        ui.Service
	config    *config.Config
	yt        *yt.Service
	gemini    *gemini.Service
}

func New(
	postsRepo *postsRepo.Repository,
	rdb *rdb.Service,
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
