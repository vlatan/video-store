package user

import (
	"github.com/vlatan/video-store/internal/avatar"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/integrations/r2"
	usersRepo "github.com/vlatan/video-store/internal/repos/users"
	"github.com/vlatan/video-store/internal/tmpl"
)

type Service struct {
	usersRepo *usersRepo.Repository
	avatar    *avatar.Service
	rdb       *rdb.Service
	r2s       *r2.Service
	tmpl      *tmpl.Service
	config    *config.Config
}

func New(
	usersRepo *usersRepo.Repository,
	avatar *avatar.Service,
	rdb *rdb.Service,
	r2s *r2.Service,
	tmpl *tmpl.Service,
	config *config.Config,
) *Service {
	return &Service{
		usersRepo: usersRepo,
		avatar:    avatar,
		rdb:       rdb,
		r2s:       r2s,
		tmpl:      tmpl,
		config:    config,
	}
}
