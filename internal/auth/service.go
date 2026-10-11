package auth

import (
	"context"

	"github.com/vlatan/video-store/internal/avatar"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/store"
	"github.com/vlatan/video-store/internal/tmpl"
	"github.com/vlatan/video-store/internal/types"
)

type usersRepo interface {
	UpsertUser(context.Context, *types.User) (int, error)
	DeleteUser(context.Context, int) (int64, error)
}

type Service struct {
	repo      usersRepo
	avatar    *avatar.Service
	session   *store.Service
	tmpl      *tmpl.Service
	config    *config.Config
	providers Providers
}

func New(
	repo usersRepo,
	avatar *avatar.Service,
	store *store.Service,
	tmpl *tmpl.Service,
	config *config.Config,
) *Service {
	return &Service{
		repo:      repo,
		avatar:    avatar,
		session:   store,
		tmpl:      tmpl,
		config:    config,
		providers: NewProviders(config),
	}
}
