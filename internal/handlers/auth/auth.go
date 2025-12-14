package auth

import (
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/redis"
	"github.com/vlatan/video-store/internal/integrations/r2"
	"github.com/vlatan/video-store/internal/repositories/users"
	"github.com/vlatan/video-store/internal/ui"

	"github.com/gorilla/sessions"
)

type Service struct {
	usersRepo *users.Repository
	store     sessions.Store
	rdb       *redis.RedisService
	r2s       r2.Service
	ui        ui.Service
	config    *config.Config
	providers Providers
}

func New(
	usersRepo *users.Repository,
	store sessions.Store,
	rdb *redis.RedisService,
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
