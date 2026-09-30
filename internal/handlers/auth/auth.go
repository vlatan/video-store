package auth

import (
	"github.com/vlatan/video-store/internal/avatars"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/integrations/r2"
	"github.com/vlatan/video-store/internal/repos/users"
	"github.com/vlatan/video-store/internal/store"
	"github.com/vlatan/video-store/internal/ui"
)

type Service struct {
	usersRepo *users.Repository
	avatars   *avatars.Service
	session   *store.Service
	rdb       *rdb.Service
	r2s       r2.Service
	ui        ui.Service
	config    *config.Config
	providers Providers
}

func New(
	usersRepo *users.Repository,
	avatars *avatars.Service,
	store *store.Service,
	rdb *rdb.Service,
	r2s r2.Service,
	ui ui.Service,
	config *config.Config,
) *Service {
	return &Service{
		usersRepo: usersRepo,
		avatars:   avatars,
		session:   store,
		rdb:       rdb,
		r2s:       r2s,
		ui:        ui,
		config:    config,
		providers: NewProviders(config),
	}
}
