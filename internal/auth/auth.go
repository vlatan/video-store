package auth

import (
	"factual-docs/internal/services/config"
	"factual-docs/internal/services/redis"
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
