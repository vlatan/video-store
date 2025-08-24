package auth

import (
	"factual-docs/internal/config"
	"factual-docs/internal/drivers/redis"
	"factual-docs/internal/r2"
	"factual-docs/internal/repositories/users"
	"factual-docs/internal/ui"

	"github.com/gorilla/sessions"
)

type Service struct {
	usersRepo *users.Repository
	store     sessions.Store
	rdb       redis.Service
	r2s       r2.Service
	ui        ui.Service
	config    *config.Config
	providers Providers
}

func New(
	usersRepo *users.Repository,
	store sessions.Store,
	rdb redis.Service,
	r2s r2.Service,
	ui ui.Service,
	config *config.Config,
) *Service {
	return &Service{
		usersRepo: usersRepo,
		store:     store,
		rdb:       rdb,
		r2s:       r2s,
		ui:        ui,
		config:    config,
		providers: NewProviders(config),
	}
}
