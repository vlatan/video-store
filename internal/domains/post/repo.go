package post

import (
	"embed"
	"io/fs"
	"text/template"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/database"
)

//go:embed sql/*.sql
var sqlFS embed.FS

type Repository struct {
	cfg  *config.Config
	db   *database.Service
	tmpl *template.Template
}

func New(cfg *config.Config, db *database.Service, fsys fs.FS) (*Repository, error) {

	if fsys == nil {
		fsys = sqlFS
	}

	tmpl, err := template.ParseFS(fsys, "sql/*.sql")
	if err != nil {
		return nil, err
	}

	return &Repository{cfg, db, tmpl}, nil
}
