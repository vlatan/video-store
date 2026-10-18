package auth

import (
	"context"

	"github.com/vlatan/video-store/internal/avatar"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/domains/user"
	"github.com/vlatan/video-store/internal/drivers/store"
	"github.com/vlatan/video-store/internal/tmpl"
)

type upserter interface {
	Upsert(context.Context, *user.User) (int, error)
	Delete(context.Context, int) (int64, error)
}

type Service struct {
	cfg       *config.Config
	user      upserter
	avatar    *avatar.Service
	session   *store.Service
	tmpl      *tmpl.Service
	providers Providers
}

func New(
	cfg *config.Config,
	user upserter,
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
