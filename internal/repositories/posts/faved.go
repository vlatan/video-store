package posts

import (
	"context"
	"encoding/json"
	"factual-docs/internal/models"
	"fmt"
)

const getUserFavedPostsQuery = `
	SELECT
		p.video_id,
		p.title,
		p.thumbnails,
		COUNT(pl.id) AS likes,
		CASE 
			WHEN $3 = 0 THEN COUNT(*) OVER()
			ELSE 0
		END AS total_results
	FROM post AS p
	LEFT JOIN post_like AS pl ON pl.post_id = p.id
	LEFT JOIN post_fave AS pf ON pf.post_id = p.id
	WHERE pf.user_id = $1
	GROUP BY p.id, pf.id
	ORDER BY pf.created_at, p.upload_date
	LIMIT $2 OFFSET $3
`

// Get user's favorited posts
func (r *Repository) GetUserFavedPosts(ctx context.Context, userID, page int) (*models.Posts, error) {

	var posts models.Posts

	// Construct the limit and offset
	limit := r.config.PostsPerPage
	offset := (page - 1) * limit

	// Get rows from DB
	rows, err := r.db.Query(ctx, getUserFavedPostsQuery, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	for rows.Next() {
		var post models.Post
		var thumbnails []byte
		var totalNum int

		// Paste post from row to struct, thumbnails in a separate var
		if err = rows.Scan(&post.VideoID, &post.Title, &thumbnails, &post.Likes, &totalNum); err != nil {
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
		posts.Items = append(posts.Items, post)
		if totalNum != 0 {
			posts.TotalNum = totalNum
		}
	}

	// If error during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &posts, nil
}
