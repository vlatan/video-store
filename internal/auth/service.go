package auth

import (
	"context"

	"github.com/vlatan/video-store/internal/avatar"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/store"
	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/ui"
)

type usersRepo interface {
	UpsertUser(context.Context, *types.User) (int, error)
	DeleteUser(context.Context, int) (int64, error)
}

type Service struct {
	repo      usersRepo
	avatar    *avatar.Service
	session   *store.Service
	ui        *ui.Service
	config    *config.Config
	providers Providers
}

func New(
	repo usersRepo,
	avatar *avatar.Service,
	store *store.Service,
	ui *ui.Service,
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
