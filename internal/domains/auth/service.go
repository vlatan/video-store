package auth

import (
	"context"

	"github.com/vlatan/video-store/internal/avatar"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/store"
	"github.com/vlatan/video-store/internal/tmpl"
	"github.com/vlatan/video-store/internal/types"
)

type userRepo interface {
	Upsert(context.Context, *types.User) (int, error)
	Delete(context.Context, int) (int64, error)
}

type Service struct {
	cfg       *config.Config
	user      userRepo
	avatar    *avatar.Service
	session   *store.Service
	tmpl      *tmpl.Service
	providers Providers
}

func New(
	cfg *config.Config,
	user userRepo,
	avatar *avatar.Service,
	store *store.Service,
	tmpl *tmpl.Service,
) *Service {
	return &Service{
		cfg:       cfg,
		user:      user,
		avatar:    avatar,
		session:   store,
		tmpl:      tmpl,
		providers: NewProviders(cfg),
	}
}
