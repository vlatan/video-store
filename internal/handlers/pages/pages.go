package pages

import (
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	pagesRepo "github.com/vlatan/video-store/internal/repositories/pages"
	"github.com/vlatan/video-store/internal/ui"
)

type Service struct {
	pagesRepo *pagesRepo.Repository
	rdb       *rdb.Service
	ui        ui.Service
	config    *config.Config
}

func New(
	pagesRepo *pagesRepo.Repository,
	rdb *rdb.Service,
	ui ui.Service,
	config *config.Config,
) *Service {
	return &Service{
		pagesRepo: pagesRepo,
		rdb:       rdb,
		ui:        ui,
		config:    config,
	}
}
