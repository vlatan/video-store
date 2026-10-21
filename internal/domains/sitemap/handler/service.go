package handler

import (
	"fmt"
	"html/template"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/domains/sitemap"
	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/tmpl"
)

const (
	sitemapPartsNum = 20
	sitemapRedisKey = "sitemap:data"
)

var sitemapPartTypes = []string{
	"post",
	"misc",
}

type Service struct {
	cfg     *config.Config
	sitemap *sitemap.Repository
	rdb     *rdb.Service
	tmpl    *tmpl.Service
	sqlArgs []any
}

type tmplData struct {
	XMLDeclarations []template.HTML
	SitemapItems    []*sitemap.Item
	tmpl.Base
}

func New(
	cfg *config.Config,
	db *database.Service,
	rdb *rdb.Service,
	tmpl *tmpl.Service,
) (*Service, error) {

	repo, err := sitemap.NewRepo(db, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create sitemap repo: %w", err)
	}

	args := make([]any, 0, 1+len(sitemapPartTypes))
	args = append(args, sitemapPartsNum)
	for _, t := range sitemapPartTypes {
		args = append(args, t)
	}

	return &Service{
		cfg:     cfg,
		sitemap: repo,
		rdb:     rdb,
		tmpl:    tmpl,
		sqlArgs: args,
	}, nil
}
