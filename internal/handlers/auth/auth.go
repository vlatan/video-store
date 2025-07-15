package auth

import (
	"factual-docs/internal/repositories/users"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/redis"
	tmpls "factual-docs/internal/shared/templates"

	"github.com/gorilla/sessions"
)

type Service struct {
	usersRepo *users.Repository
	store     *sessions.CookieStore
	rdb       redis.Service
	tm        tmpls.Service
	config    *config.Config
}

func New(
	usersRepo *users.Repository,
	store *sessions.CookieStore,
	rdb redis.Service,
	tm tmpls.Service,
	config *config.Config,
) *Service {
	return &Service{
		usersRepo: usersRepo,
		store:     store,
		rdb:       rdb,
		tm:        tm,
		config:    config,
	}
}
