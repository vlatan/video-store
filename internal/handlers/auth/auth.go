package auth

import (
	"factual-docs/internal/config"
	"factual-docs/internal/drivers/redis"
	"factual-docs/internal/oauth"
	"factual-docs/internal/repositories/users"
	"factual-docs/internal/ui"

	"github.com/gorilla/sessions"
)

type Service struct {
	usersRepo *users.Repository
	store     sessions.Store
	rdb       redis.Service
	ui        ui.Service
	config    *config.Config
	oauth     oauth.Providers
}

func New(
	usersRepo *users.Repository,
	store sessions.Store,
	rdb redis.Service,
	ui ui.Service,
	config *config.Config,
	oAuth oauth.Providers,
) *Service {
	return &Service{
		usersRepo: usersRepo,
		store:     store,
		rdb:       rdb,
		ui:        ui,
		config:    config,
		oauth:     oAuth,
	}
}
