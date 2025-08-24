package users

import (
	"factual-docs/internal/config"
	"factual-docs/internal/drivers/redis"
	"factual-docs/internal/r2"
	postsRepo "factual-docs/internal/repositories/posts"
	usersRepo "factual-docs/internal/repositories/users"
	"factual-docs/internal/ui"
)

type Service struct {
	usersRepo *usersRepo.Repository
	postsRepo *postsRepo.Repository
	rdb       redis.Service
	r2s       r2.Service
	ui        ui.Service
	config    *config.Config
}

func New(
	usersRepo *usersRepo.Repository,
	postsRepo *postsRepo.Repository,
	rdb redis.Service,
	r2s r2.Service,
	ui ui.Service,
	config *config.Config,
) *Service {
	return &Service{
		usersRepo: usersRepo,
		postsRepo: postsRepo,
		rdb:       rdb,
		r2s:       r2s,
		ui:        ui,
		config:    config,
	}
}
