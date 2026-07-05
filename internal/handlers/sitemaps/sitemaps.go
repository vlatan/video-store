package sitemaps

import (
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	postsRepo "github.com/vlatan/video-store/internal/repositories/posts"
	"github.com/vlatan/video-store/internal/ui"
)

const (
	sitemapPartsNum = 20
	sitemapRedisKey = "sitemap:data"
)

var sitemapTypes = []string{
	"post",
	"misc",
}

type Service struct {
	postsRepo *postsRepo.Repository
	rdb       *rdb.Service
	ui        ui.Service
	config    *config.Config
	sqlArgs   []any
}

func New(
	postsRepo *postsRepo.Repository,
	rdb *rdb.Service,
	ui ui.Service,
	config *config.Config,
) *Service {

	args := make([]any, 0, 1+len(sitemapTypes))
	args = append(args, sitemapPartsNum)
	for _, t := range sitemapTypes {
		args = append(args, t)
	}

	return &Service{
		postsRepo: postsRepo,
		rdb:       rdb,
		ui:        ui,
		config:    config,
		sqlArgs:   args,
	}
}
