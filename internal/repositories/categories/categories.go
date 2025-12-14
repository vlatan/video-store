package categories

import (
	"context"

	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/models"
)

type Repository struct {
	db *database.Service
}

func New(db *database.Service) *Repository {
	return &Repository{db: db}
}

// Get all valid categories
func (r *Repository) GetCategories(ctx context.Context) (models.Categories, error) {

	rows, err := r.db.Pool.Query(ctx, getCategoriesQuery)
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
