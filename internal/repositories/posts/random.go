package posts

import (
	"context"
	"database/sql"
	"errors"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

const getRandomPostsQuery = `
	SELECT
		p.video_id,
		p.title,
		p.original_title,
		p.thumbnails,
		COUNT(pl.id) AS likes
	FROM post AS p
	LEFT JOIN post_like AS pl ON pl.post_id = p.id
	WHERE p.title != $1 AND p.original_title != $1
	GROUP BY p.id
	ORDER BY RANDOM()
	LIMIT $2
`

// Get random posts, but exclude posts with the exact title match
func (r *Repository) GetRandomPosts(ctx context.Context, title string, limit int) (models.Posts, error) {

	var zero, posts models.Posts

	if title == "" {
		return zero, errors.New("title can't be empty string")
	}

	// Get rows from DB
	rows, err := r.db.Pool.Query(ctx, getRandomPostsQuery, title, limit)
	if err != nil {
		return zero, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	for rows.Next() {
		var post models.Post
		var originalTitle sql.NullString

		if err = rows.Scan(
			&post.VideoID,
			&post.Title,
			&originalTitle,
			&post.RawThumbs,
			&post.Likes,
		); err != nil {
			return zero, err
		}

		// Include the processed post in the result
		post.OriginalTitle = utils.FromNullString(originalTitle)
		posts.Items = append(posts.Items, post)
	}

	// If error during iteration
	if err = rows.Err(); err != nil {
		return zero, err
	}

	// Post-process the posts, prepare the thumbnail
	if err = postProcessPosts(ctx, posts); err != nil {
		return zero, err
	}

	return posts, nil
}
