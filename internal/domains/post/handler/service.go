package handler

import (
	"github.com/vlatan/video-store/internal/avatar"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/domains/post"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/drivers/store"
	"github.com/vlatan/video-store/internal/integrations/gemini"
	"github.com/vlatan/video-store/internal/integrations/yt"
	"github.com/vlatan/video-store/internal/tmpl"
)

type Service struct {
	cfg     *config.Config
	post    *post.Repository
	rdb     *rdb.Service
	session *store.Service
	avatar  *avatar.Service
	tmpl    *tmpl.Service
	yt      *yt.Service
	gemini  *gemini.Service
}

type tmplData struct {
	Title       string
	CurrentPost *post.Post
	post.Posts
	tmpl.Base
	Form
}

func New(
	cfg *config.Config,
	post *post.Repository,
	rdb *rdb.Service,
	store *store.Service,
	avatar *avatar.Service,
	tmpl *tmpl.Service,
	yt *yt.Service,
	gemini *gemini.Service,
) *Service {
	return &Service{
		cfg:     cfg,
		post:    post,
		rdb:     rdb,
		session: store,
		avatar:  avatar,
		tmpl:    tmpl,
		yt:      yt,
		gemini:  gemini,
	}
}
