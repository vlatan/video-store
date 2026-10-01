package pages

import (
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	pagesRepo "github.com/vlatan/video-store/internal/repos/pages"
	"github.com/vlatan/video-store/internal/store"
	"github.com/vlatan/video-store/internal/ui"
)

type Service struct {
	pagesRepo *pagesRepo.Repository
	rdb       *rdb.Service
	session   *store.Service
	ui        ui.Service
	config    *config.Config
}

func New(
	pagesRepo *pagesRepo.Repository,
	rdb *rdb.Service,
	store *store.Service,
	ui ui.Service,
	config *config.Config,
) *Service {
	return &Service{
		pagesRepo: pagesRepo,
		rdb:       rdb,
		session:   store,
		ui:        ui,
		config:    config,
	}
}
