package posts

import (
	"context"

	"github.com/vlatan/video-store/internal/models"
)

const getCategoryPostsQuery = `
	WITH posts_with_likes AS (
		SELECT 
			c.name AS category_title,
			post.id,
			post.video_id, 
			post.title, 
			post.thumbnails,
			COUNT(pl.id) AS likes,
			post.upload_date
		FROM post
		JOIN category AS c ON c.id = post.category_id 
		LEFT JOIN post_like AS pl ON pl.post_id = post.id
		WHERE c.slug = $1
		GROUP BY c.id, post.id
	)
	SELECT * FROM posts_with_likes
	%s --- the WHERE clause
	ORDER BY %s
	LIMIT $2
`

// Get a limited number of posts from one category with cursor
func (r *Repository) GetCategoryPosts(
	ctx context.Context,
	categorySlug,
	cursor,
	orderBy string,
) (models.Posts, error) {

	posts, err := r.queryTaxonomyPosts(
		ctx,
		getCategoryPostsQuery,
		categorySlug,
		cursor,
		orderBy,
	)

	if err != nil {
		return models.Posts{}, err
	}

	return posts, nil
}

const getSourcePostsQuery = `
	WITH posts_with_likes AS (
		SELECT 
			p.title AS playlist_title,
			post.id,
			post.video_id, 
			post.title, 
			post.thumbnails,
			COUNT(pl.id) AS likes,
			post.upload_date
		FROM post
		LEFT JOIN playlist AS p ON p.id = post.playlist_db_id 
		LEFT JOIN post_like AS pl ON pl.post_id = post.id
		WHERE
			CASE 
				WHEN $1 = 'other'
				THEN (p.playlist_id IS NULL OR p.playlist_id = '')
				ELSE p.playlist_id = $1
			END
		GROUP BY p.id, post.id
	)
	SELECT * FROM posts_with_likes
	%s --- the WHERE clause
	ORDER BY %s
	LIMIT $2
`

// Get a limited number of posts from one category with cursor
func (r *Repository) GetSourcePosts(
	ctx context.Context,
	playlistID,
	cursor,
	orderBy string,
) (models.Posts, error) {

	posts, err := r.queryTaxonomyPosts(
		ctx,
		getSourcePostsQuery,
		playlistID,
		cursor,
		orderBy,
	)

	if err != nil {
		return models.Posts{}, err
	}

	return posts, nil
}
