package ui

import (
	"html/template"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/drivers/store"
	"github.com/vlatan/video-store/internal/repos/categories"
)

type versioner interface {
	AddVersion(path string) string
}

type templateMap map[string]*template.Template

type Service struct {
	cfg       *config.Config
	rdb       *rdb.Service
	session   *store.Service
	catsRepo  *categories.Repository
	versioner versioner
	templates templateMap
}
