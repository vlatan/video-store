package posts

import (
	"github.com/vlatan/video-store/internal/avatar"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/integrations/gemini"
	"github.com/vlatan/video-store/internal/integrations/yt"
	postsRepo "github.com/vlatan/video-store/internal/repos/posts"
	usersRepo "github.com/vlatan/video-store/internal/repos/users"
	"github.com/vlatan/video-store/internal/store"
	"github.com/vlatan/video-store/internal/ui"
)

type Service struct {
	postsRepo *postsRepo.Repository
	usersRepo *usersRepo.Repository
	avatar    *avatar.Service
	rdb       *rdb.Service
	session   *store.Service
	ui        ui.Service
	config    *config.Config
	yt        *yt.Service
	gemini    *gemini.Service
}

func New(
	postsRepo *postsRepo.Repository,
	usersRepo *usersRepo.Repository,
	avatar *avatar.Service,
	rdb *rdb.Service,
	store *store.Service,
	ui ui.Service,
	config *config.Config,
	yt *yt.Service,
	gemini *gemini.Service,
) *Service {
	return &Service{
		postsRepo: postsRepo,
		usersRepo: usersRepo,
		avatar:    avatar,
		rdb:       rdb,
		session:   store,
		ui:        ui,
		config:    config,
		yt:        yt,
		gemini:    gemini,
	}
}
