package user

import (
	"github.com/vlatan/video-store/internal/avatar"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/store"
	"github.com/vlatan/video-store/internal/tmpl"
	"github.com/vlatan/video-store/internal/user/repo"
)

type Service struct {
	cfg     *config.Config
	repo    *repo.Repository
	avatar  *avatar.Service
	session *store.Service
	tmpl    *tmpl.Service
}

func New(
	cfg *config.Config,
	repo *repo.Repository,
	avatar *avatar.Service,
	store *store.Service,
	tmpl *tmpl.Service,
) *Service {
	return &Service{
		cfg:     cfg,
		repo:    repo,
		avatar:  avatar,
		session: store,
		tmpl:    tmpl,
	}
}
