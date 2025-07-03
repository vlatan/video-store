package auth

import (
	"factual-docs/internal/repositories/users"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/redis"

	"github.com/gorilla/sessions"
)

type Service struct {
	usersRepo *users.Repository
	store     *sessions.CookieStore
	rdb       redis.Service
	config    *config.Config
}

func New(usersRepo *users.Repository, store *sessions.CookieStore, rdb redis.Service, config *config.Config) *Service {
	return &Service{
		usersRepo: usersRepo,
		store:     store,
		rdb:       rdb,
		config:    config,
	}
}
