package sources

import (
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/drivers/store"
	"github.com/vlatan/video-store/internal/integrations/yt"
	postsRepo "github.com/vlatan/video-store/internal/repos/posts"
	sourcesRepo "github.com/vlatan/video-store/internal/repos/sources"
	"github.com/vlatan/video-store/internal/tmpl"
)

type Service struct {
	config      *config.Config
	rdb         *rdb.Service
	session     *store.Service
	postsRepo   *postsRepo.Repository
	sourcesRepo *sourcesRepo.Repository
	tmpl        *tmpl.Service
	yt          *yt.Service
}

func New(
	config *config.Config,
	rdb *rdb.Service,
	store *store.Service,
	postsRepo *postsRepo.Repository,
	sourcesRepo *sourcesRepo.Repository,
	tmpl *tmpl.Service,
	yt *yt.Service,
) *Service {
	return &Service{
		config:      config,
		rdb:         rdb,
		session:     store,
		postsRepo:   postsRepo,
		sourcesRepo: sourcesRepo,
		tmpl:        tmpl,
		yt:          yt,
	}
}
