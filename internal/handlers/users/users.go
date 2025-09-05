package users

import (
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/redis"
	"github.com/vlatan/video-store/internal/r2"
	postsRepo "github.com/vlatan/video-store/internal/repositories/posts"
	usersRepo "github.com/vlatan/video-store/internal/repositories/users"
	"github.com/vlatan/video-store/internal/ui"
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
