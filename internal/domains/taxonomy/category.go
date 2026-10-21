package taxonomy

import (
	"context"

	"github.com/vlatan/video-store/internal/utils/stringx"
)

// Get all valid categories
func (r *Repository) Categories(ctx context.Context) (Categories, error) {

	query, err := stringx.GetQuery(r.tmpl, "all_categories.sql", nil)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories Categories
	for rows.Next() {

		// Get categories from DB
		var category Category
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
