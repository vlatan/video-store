package posts

import (
	"context"
	"encoding/json"
	"factual-docs/internal/models"
	"fmt"
)

const getRandomPostsQuery = `
	SELECT
		p.video_id,
		p.title,
		p.thumbnails,
		COUNT(pl.id) AS likes
	FROM post AS p
	LEFT JOIN post_like AS pl ON pl.post_id = p.id
	WHERE p.title != $1
	GROUP BY p.id
	ORDER BY RANDOM()
	LIMIT $2
`

// Get random posts
func (r *Repository) GetRandomPosts(ctx context.Context, title string, limit int) ([]models.Post, error) {

	var posts []models.Post

	// Get rows from DB
	rows, err := r.db.Query(ctx, getRandomPostsQuery, title, limit)
	if err != nil {
		return nil, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	for rows.Next() {
		var post models.Post
		var thumbnails []byte

		// Paste post from row to struct, thumbnails in a separate var
		if err = rows.Scan(&post.VideoID, &post.Title, &thumbnails, &post.Likes); err != nil {
			return nil, err
		}

		// Unserialize thumbnails
		var thumbs models.Thumbnails
		if err = json.Unmarshal(thumbnails, &thumbs); err != nil {
			return nil, fmt.Errorf("video ID '%s': %w", post.VideoID, err)
		}

		// Craft srcset string
		post.Srcset = thumbs.Srcset(480)
		post.Thumbnail = thumbs.Medium

		// Include the processed post in the result
		posts = append(posts, post)
	}

	// If error during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}
