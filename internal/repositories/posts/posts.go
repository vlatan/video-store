package posts

import (
	"embed"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/database"
)

type Repository struct {
	db     *database.Service
	config *config.Config
}

func New(db *database.Service, config *config.Config) *Repository {
	return &Repository{
		db:     db,
		config: config,
	}
}

//go:embed *.sql
var localQueries embed.FS

// Packaged private query cache
// var queryCache *sqlutils.Cache
