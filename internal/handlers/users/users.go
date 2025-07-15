package users

import (
	"factual-docs/internal/handlers/auth"
	postsRepo "factual-docs/internal/repositories/posts"
	usersRepo "factual-docs/internal/repositories/users"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/redis"
	tmpls "factual-docs/internal/shared/templates"
)

type Service struct {
	usersRepo *usersRepo.Repository
	postsRepo *postsRepo.Repository
	rdb       redis.Service
	tm        tmpls.Service
	config    *config.Config
	auth      *auth.Service
}

func New(
	usersRepo *usersRepo.Repository,
	postsRepo *postsRepo.Repository,
	rdb redis.Service,
	tm tmpls.Service,
	config *config.Config,
	auth *auth.Service,
) *Service {
	return &Service{
		usersRepo: usersRepo,
		postsRepo: postsRepo,
		rdb:       rdb,
		tm:        tm,
		config:    config,
		auth:      auth,
	}
}
