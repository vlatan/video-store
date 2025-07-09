package sources

import (
	"context"
	"factual-docs/internal/models"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/database"
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

// Get a limited number of posts with offset
func (r *Repository) GetSources(ctx context.Context) ([]models.Source, error) {

	rows, err := r.db.Query(ctx, getSourcesQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []models.Source
	for rows.Next() {

		// Get categories from DB
		var source models.Source
		var thumbnails []byte
		if err := rows.Scan(&source.PlaylistID, &source.Title, &thumbnails); err != nil {
			return []models.Source{}, err
		}

		// Include the category in the result
		sources = append(sources, source)
	}

	if err := rows.Err(); err != nil {
		return []models.Source{}, err
	}

	return sources, nil
}
