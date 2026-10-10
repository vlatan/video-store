package posts

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
	db     *database.Service
	config *config.Config
	tmpl   *template.Template
}

func New(db *database.Service, config *config.Config, fsys fs.FS) (*Repository, error) {

	if fsys == nil {
		fsys = sqlFS
	}

	tmpl, err := template.ParseFS(fsys, "sql/*.sql")
	if err != nil {
		return nil, err
	}

	return &Repository{db, config, tmpl}, nil
}
