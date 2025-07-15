package auth

import (
	"factual-docs/internal/repositories/users"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/redis"
	"factual-docs/internal/shared/view"

	"github.com/gorilla/sessions"
)

type Service struct {
	usersRepo *users.Repository
	store     *sessions.CookieStore
	rdb       redis.Service
	view      view.Service
	config    *config.Config
}

func New(
	usersRepo *users.Repository,
	store *sessions.CookieStore,
	rdb redis.Service,
	view view.Service,
	config *config.Config,
) *Service {
	return &Service{
		usersRepo: usersRepo,
		store:     store,
		rdb:       rdb,
		view:      view,
		config:    config,
	}
}
