package users

import (
	"factual-docs/internal/config"
	postsRepo "factual-docs/internal/repositories/posts"
	usersRepo "factual-docs/internal/repositories/users"
	"factual-docs/internal/shared/redis"
	"factual-docs/internal/ui"
)

type Service struct {
	usersRepo *usersRepo.Repository
	postsRepo *postsRepo.Repository
	rdb       redis.Service
	ui        ui.Service
	config    *config.Config
}

func New(
	usersRepo *usersRepo.Repository,
	postsRepo *postsRepo.Repository,
	rdb redis.Service,
	ui ui.Service,
	config *config.Config,
) *Service {
	return &Service{
		usersRepo: usersRepo,
		postsRepo: postsRepo,
		rdb:       rdb,
		ui:        ui,
		config:    config,
	}
}
