package posts

import (
	"factual-docs/internal/config"
	"factual-docs/internal/drivers/database"
)

type Repository struct {
	db     database.Service
	config *config.Config
}

func New(db database.Service, config *config.Config) *Repository {
	return &Repository{
		db:     db,
		config: config,
	}
}
