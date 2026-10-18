package handler

import (
	"context"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/domains/post"
	"github.com/vlatan/video-store/internal/domains/taxonomy"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/drivers/store"
	"github.com/vlatan/video-store/internal/integrations/yt"
	"github.com/vlatan/video-store/internal/tmpl"
	"github.com/vlatan/video-store/internal/types"
)

type poster interface {
	Category(context.Context, string, string, string) (post.Posts, error)
	Source(context.Context, string, string, string) (post.Posts, error)
	Year(context.Context, string, string, string) (post.Posts, error)
}

type Service struct {
	cfg      *config.Config
	taxonomy *taxonomy.Repository
	post     poster
	rdb      *rdb.Service
	session  *store.Service
	tmpl     *tmpl.Service
	yt       *yt.Service
}

type tmplData struct {
	Title           string
	CurrentYear     int16
	CurrentCategory *taxonomy.Category
	CurrentSource   *taxonomy.Source
	Sources         []taxonomy.Source
	types.Form
	post.Posts
	tmpl.Base
}

func New(
	cfg *config.Config,
	taxonomy *taxonomy.Repository,
	post poster,
	rdb *rdb.Service,
	store *store.Service,
	tmpl *tmpl.Service,
	yt *yt.Service,
) *Service {
	return &Service{
		cfg:      cfg,
		taxonomy: taxonomy,
		post:     post,
		rdb:      rdb,
		session:  store,
		tmpl:     tmpl,
		yt:       yt,
	}
}
