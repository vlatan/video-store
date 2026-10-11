package posts

import (
	"github.com/vlatan/video-store/internal/avatar"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/drivers/store"
	"github.com/vlatan/video-store/internal/integrations/gemini"
	"github.com/vlatan/video-store/internal/integrations/yt"
	postsRepo "github.com/vlatan/video-store/internal/repos/posts"
	"github.com/vlatan/video-store/internal/tmpl"
)

type Service struct {
	postsRepo *postsRepo.Repository
	avatar    *avatar.Service
	rdb       *rdb.Service
	session   *store.Service
	tmpl      *tmpl.Service
	config    *config.Config
	yt        *yt.Service
	gemini    *gemini.Service
}

func New(
	postsRepo *postsRepo.Repository,
	avatar *avatar.Service,
	rdb *rdb.Service,
	store *store.Service,
	tmpl *tmpl.Service,
	config *config.Config,
	yt *yt.Service,
	gemini *gemini.Service,
) *Service {
	return &Service{
		postsRepo: postsRepo,
		avatar:    avatar,
		rdb:       rdb,
		session:   store,
		tmpl:      tmpl,
		config:    config,
		yt:        yt,
		gemini:    gemini,
	}
}
