package user

import (
	"fmt"

	"github.com/vlatan/video-store/internal/avatar"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/tmpl"
	"github.com/vlatan/video-store/internal/user/repo"
)

type Service struct {
	cfg    *config.Config
	Repo   *repo.Repository
	avatar *avatar.Service
	tmpl   *tmpl.Service
}

func New(
	cfg *config.Config,
	db *database.Service,
	avatar *avatar.Service,
	tmpl *tmpl.Service,
) (*Service, error) {

	repo, err := repo.New(cfg, db, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create user repo: %w", err)
	}

	return &Service{
		cfg:    cfg,
		Repo:   repo,
		avatar: avatar,
		tmpl:   tmpl,
	}, nil
}
