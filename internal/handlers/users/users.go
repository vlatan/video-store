package users

import (
	postsRepo "factual-docs/internal/repositories/posts"
	usersRepo "factual-docs/internal/repositories/users"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/redis"
	"factual-docs/internal/shared/view"
)

type Service struct {
	usersRepo *usersRepo.Repository
	postsRepo *postsRepo.Repository
	rdb       redis.Service
	view      view.Service
	config    *config.Config
}

func New(
	usersRepo *usersRepo.Repository,
	postsRepo *postsRepo.Repository,
	rdb redis.Service,
	view view.Service,
	config *config.Config,
) *Service {
	return &Service{
		usersRepo: usersRepo,
		postsRepo: postsRepo,
		rdb:       rdb,
		view:      view,
		config:    config,
	}
}
