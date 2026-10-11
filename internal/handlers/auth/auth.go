package auth

import (
	"context"

	"github.com/vlatan/video-store/internal/avatar"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/store"
	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/ui"
)

type userRepo interface {
	UpsertUser(context.Context, *types.User) (int, error)
	DeleteUser(context.Context, int) (int64, error)
}

type Service struct {
	repo      userRepo
	avatar    *avatar.Service
	session   *store.Service
	ui        ui.Service
	config    *config.Config
	providers Providers
}

func New(
	repo userRepo,
	avatar *avatar.Service,
	store *store.Service,
	ui ui.Service,
	config *config.Config,
) *Service {
	return &Service{
		repo:      repo,
		avatar:    avatar,
		session:   store,
		ui:        ui,
		config:    config,
		providers: NewProviders(config),
	}
}
