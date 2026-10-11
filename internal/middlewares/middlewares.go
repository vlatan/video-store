package middlewares

import (
	"github.com/vlatan/video-store/internal/avatar"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/store"
	"github.com/vlatan/video-store/internal/repos/users"
	"github.com/vlatan/video-store/internal/tmpl"
)

type Service struct {
	config    *config.Config
	session   *store.Service
	usersRepo *users.Repository
	avatar    *avatar.Service
	tmpl      *tmpl.Service
}

// New creates new middlewares service
func New(
	config *config.Config,
	store *store.Service,
	usersRepo *users.Repository,
	avatar *avatar.Service,
	tmpl *tmpl.Service) *Service {
	SetCustomLogger(config)
	return &Service{
		config:    config,
		session:   store,
		usersRepo: usersRepo,
		avatar:    avatar,
		tmpl:      tmpl,
	}
}
