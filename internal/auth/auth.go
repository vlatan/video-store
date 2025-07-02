package auth

import (
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/redis"
	"factual-docs/internal/users"

	"github.com/gorilla/sessions"
)

type Service struct {
	users  *users.Service
	store  *sessions.CookieStore
	rdb    redis.Service
	config *config.Config
}

func New(users *users.Service, store *sessions.CookieStore, rdb redis.Service, config *config.Config) *Service {
	return &Service{
		users:  users,
		store:  store,
		rdb:    rdb,
		config: config,
	}
}
