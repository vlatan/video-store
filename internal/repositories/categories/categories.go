package categories

import (
	"context"
	"factual-docs/internal/models"
	"factual-docs/internal/shared/database"
)

type Repository struct {
	db database.Service
}

func New(db database.Service) *Repository {
	return &Repository{db: db}
}

// Get a limited number of posts with offset
func (r *Repository) GetCategories(ctx context.Context) ([]models.Category, error) {

	rows, err := r.db.Query(ctx, getCategoriesQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {

		// Get categories from DB
		var category models.Category
		if err := rows.Scan(&category.Name, &category.Slug); err != nil {
			return []models.Category{}, err
		}

		// Include the category in the result
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return []models.Category{}, err
	}

	return categories, nil
}
