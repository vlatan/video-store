package sitemap

import (
	"fmt"

	"github.com/vlatan/video-store/internal/config"
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
	repo    *Repository
	rdb     *rdb.Service
	tmpl    *tmpl.Service
	sqlArgs []any
}

func New(
	cfg *config.Config,
	db *database.Service,
	rdb *rdb.Service,
	tmpl *tmpl.Service,
) (*Service, error) {

	args := make([]any, 0, 1+len(sitemapPartTypes))
	args = append(args, sitemapPartsNum)
	for _, t := range sitemapPartTypes {
		args = append(args, t)
	}

	repo, err := NewRepo(db, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create sitemap repo: %w", err)
	}

	return &Service{
		cfg:     cfg,
		repo:    repo,
		rdb:     rdb,
		tmpl:    tmpl,
		sqlArgs: args,
	}, nil
}
