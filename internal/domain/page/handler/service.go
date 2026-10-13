package handler

import (
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/domain/page/repo"
	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/drivers/store"
	"github.com/vlatan/video-store/internal/tmpl"
)

type Service struct {
	cfg     *config.Config
	repo    *repo.Repository
	rdb     *rdb.Service
	session *store.Service
	tmpl    *tmpl.Service
}

func New(
	cfg *config.Config,
	db *database.Service,
	rdb *rdb.Service,
	store *store.Service,
	tmpl *tmpl.Service,
) *Service {
	return &Service{
		cfg:     cfg,
		repo:    repo.NewRepo(db),
		rdb:     rdb,
		session: store,
		tmpl:    tmpl,
	}
}
