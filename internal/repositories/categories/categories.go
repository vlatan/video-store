package categories

import (
	"context"
	"embed"
	"fmt"

	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/repositories/sqlutils"
)

type Repository struct {
	db         *database.Service
	queryCache *sqlutils.Cache
}

//go:embed sql
var localQueries embed.FS

func New(db *database.Service) (*Repository, error) {

	queryCache, err := sqlutils.LoadTemplates(localQueries, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load categories SQL queries")
	}

	return &Repository{db, queryCache}, nil
}

// Get all valid categories
func (r *Repository) GetCategories(ctx context.Context) (models.Categories, error) {

	query, err := r.queryCache.Render("all_categories.sql", nil)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories models.Categories
	for rows.Next() {

		// Get categories from DB
		var category models.Category
		if err := rows.Scan(&category.Name, &category.Slug, &category.UpdatedAt); err != nil {
			return nil, err
		}

		// Include the category in the result
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}
