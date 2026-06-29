package posts

import (
	"embed"
	"fmt"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/repositories/sqlutils"
)

type Repository struct {
	db         *database.Service
	config     *config.Config
	queryCache *sqlutils.Cache
}

//go:embed sql
var localQueries embed.FS

func New(db *database.Service, config *config.Config) (*Repository, error) {

	queryCache, err := sqlutils.LoadTemplates(localQueries, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load the sql queries")
	}

	return &Repository{db, config, queryCache}, nil
}
