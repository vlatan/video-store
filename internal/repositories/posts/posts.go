package posts

import (
	"embed"
	"io/fs"
	"text/template"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/database"
	repo "github.com/vlatan/video-store/internal/repositories"
)

//go:embed sql/*.sql
var sqlFS embed.FS

type Repository struct {
	db      *database.Service
	config  *config.Config
	queries *template.Template
}

func New(db *database.Service, config *config.Config, fsys fs.FS) (*Repository, error) {

	if fsys == nil {
		fsys = sqlFS
	}

	queries, err := template.ParseFS(fsys, "sql/*.sql")
	if err != nil {
		return nil, err
	}

	return &Repository{db, config, queries}, nil
}

func (r *Repository) GetQuery(name string, sqlParts any) (string, error) {
	return repo.GetQuery(r.queries, name, sqlParts)
}
