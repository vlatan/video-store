package repo

import (
	"context"

	"github.com/vlatan/video-store/internal/types"
)

// Get a limited number of posts from one category with cursor
func (r *Repository) YearPosts(
	ctx context.Context,
	releaseYear,
	cursor,
	orderBy string,
) (types.Posts, error) {

	return r.queryTaxonomyPosts(
		ctx,
		"year_posts.sql",
		releaseYear,
		cursor,
		orderBy,
	)
}
