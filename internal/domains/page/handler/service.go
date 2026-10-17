package handler

import (
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/domains/page"
	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/drivers/store"
	"github.com/vlatan/video-store/internal/tmpl"
	"github.com/vlatan/video-store/internal/types"
)

type Service struct {
	cfg     *config.Config
	page    *page.Repository
	rdb     *rdb.Service
	session *store.Service
	tmpl    *tmpl.Service
}

type pageData struct {
	Title       string
	CurrentPage *page.Page
	types.Form
	tmpl.Base
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
		page:    page.NewRepo(db),
		rdb:     rdb,
		session: store,
		tmpl:    tmpl,
	}
}
