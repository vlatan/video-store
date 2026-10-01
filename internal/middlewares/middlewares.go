package middlewares

import (
	"github.com/vlatan/video-store/internal/avatar"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/repos/users"
	"github.com/vlatan/video-store/internal/store"
	"github.com/vlatan/video-store/internal/ui"
)

type Service struct {
	config    *config.Config
	session   *store.Service
	usersRepo *users.Repository
	avatar    *avatar.Service
	ui        ui.Service
}

// New creates new middlewares service
func New(
	config *config.Config,
	store *store.Service,
	usersRepo *users.Repository,
	avatar *avatar.Service,
	ui ui.Service) *Service {
	SetCustomLogger(config)
	return &Service{
		config:    config,
		session:   store,
		usersRepo: usersRepo,
		avatar:    avatar,
		ui:        ui,
	}
}
