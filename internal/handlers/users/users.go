package users

import (
	"github.com/vlatan/video-store/internal/avatar"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/integrations/r2"
	postsRepo "github.com/vlatan/video-store/internal/repos/posts"
	usersRepo "github.com/vlatan/video-store/internal/repos/users"
	"github.com/vlatan/video-store/internal/ui"
)

type Service struct {
	usersRepo *usersRepo.Repository
	postsRepo *postsRepo.Repository
	avatar    *avatar.Service
	rdb       *rdb.Service
	r2s       r2.Service
	ui        ui.Service
	config    *config.Config
}

func New(
	usersRepo *usersRepo.Repository,
	postsRepo *postsRepo.Repository,
	avatar *avatar.Service,
	rdb *rdb.Service,
	r2s r2.Service,
	ui ui.Service,
	config *config.Config,
) *Service {
	return &Service{
		usersRepo: usersRepo,
		postsRepo: postsRepo,
		avatar:    avatar,
		rdb:       rdb,
		r2s:       r2s,
		ui:        ui,
		config:    config,
	}
}
