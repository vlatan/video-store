package pages

import (
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/drivers/store"
	pagesRepo "github.com/vlatan/video-store/internal/repos/pages"
	"github.com/vlatan/video-store/internal/tmpl"
)

type Service struct {
	pagesRepo *pagesRepo.Repository
	rdb       *rdb.Service
	session   *store.Service
	tmpl      *tmpl.Service
	config    *config.Config
}

func New(
	pagesRepo *pagesRepo.Repository,
	rdb *rdb.Service,
	store *store.Service,
	tmpl *tmpl.Service,
	config *config.Config,
) *Service {
	return &Service{
		pagesRepo: pagesRepo,
		rdb:       rdb,
		session:   store,
		tmpl:      tmpl,
		config:    config,
	}
}
