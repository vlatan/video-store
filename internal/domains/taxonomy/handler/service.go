package handler

import (
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/domains/taxonomy/repo"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/drivers/store"
	"github.com/vlatan/video-store/internal/integrations/yt"
	"github.com/vlatan/video-store/internal/tmpl"
	"github.com/vlatan/video-store/internal/types"
)

type Service struct {
	cfg     *config.Config
	repo    *repo.Repository
	rdb     *rdb.Service
	session *store.Service
	tmpl    *tmpl.Service
	yt      *yt.Service
}

type taxonomyData struct {
	Title           string
	CurrentYear     int16
	CurrentCategory *types.Category
	CurrentSource   *types.Source
	Sources         []types.Source
	types.Form
	types.Posts
	tmpl.Base
}

func New(
	cfg *config.Config,
	repo *repo.Repository,
	rdb *rdb.Service,
	store *store.Service,
	tmpl *tmpl.Service,
	yt *yt.Service,
) *Service {
	return &Service{
		cfg:     cfg,
		repo:    repo,
		rdb:     rdb,
		session: store,
		tmpl:    tmpl,
		yt:      yt,
	}
}
