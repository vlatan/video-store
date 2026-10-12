package middlewares

import (
	"context"

	"github.com/vlatan/video-store/internal/avatar"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/store"
	"github.com/vlatan/video-store/internal/tmpl"
)

type userRepo interface {
	UpdateLastSeen(context.Context, int) (int64, error)
}

type Service struct {
	cfg     *config.Config
	user    userRepo
	session *store.Service
	avatar  *avatar.Service
	tmpl    *tmpl.Service
}

// New creates new middlewares service
func New(
	cfg *config.Config,
	user userRepo,
	store *store.Service,
	avatar *avatar.Service,
	tmpl *tmpl.Service) *Service {
	SetCustomLogger(cfg)
	return &Service{
		cfg:     cfg,
		user:    user,
		session: store,
		avatar:  avatar,
		tmpl:    tmpl,
	}
}
