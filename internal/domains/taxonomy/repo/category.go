package repo

import (
	"context"

	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/utils/stringx"
)

// Get all valid categories
func (r *Repository) Categories(ctx context.Context) (types.Categories, error) {

	query, err := stringx.GetQuery(r.tmpl, "all_categories.sql", nil)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories types.Categories
	for rows.Next() {

		// Get categories from DB
		var category types.Category
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

// Get a limited number of posts from one category with cursor
func (r *Repository) CategoryPosts(
	ctx context.Context,
	categorySlug,
	cursor,
	orderBy string,
) (types.Posts, error) {

	return r.queryTaxonomyPosts(
		ctx,
		"category_posts.sql",
		categorySlug,
		cursor,
		orderBy,
	)
}
