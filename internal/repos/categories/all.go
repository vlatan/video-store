package categories

import (
	"context"

	"github.com/vlatan/video-store/internal/models"
)

// Get all valid categories
func (r *Repository) GetCategories(ctx context.Context) (models.Categories, error) {

	query, err := r.GetQuery("all_categories.sql", nil)
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
