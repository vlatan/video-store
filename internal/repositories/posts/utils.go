package posts

import (
	"context"
	"encoding/json"
	"factual-docs/internal/models"
	"fmt"
)

// Get post's related posts based on provided title as search query
func (r *Repository) GetRelatedPosts(ctx context.Context, title string) (posts []models.Post, err error) {
	// Search the DB for posts
	searchedPosts, err := r.SearchPosts(ctx, title, r.config.NumRelatedPosts+1, 0)

	if err != nil {
		return posts, err
	}

	for _, sp := range searchedPosts.Items {
		if sp.Title != title {
			posts = append(posts, sp)
		}
	}

	return posts, err
}

// Query the DB for posts based on variadic arguments
func (r *Repository) queryPosts(
	ctx context.Context,
	query string,
	args ...any,
) (posts []models.Post, err error) {
	// Get rows from DB
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return posts, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	for rows.Next() {
		var post models.Post
		var thumbnails []byte

		// Paste post from row to struct, thumbnails in a separate var
		if err = rows.Scan(&post.VideoID, &post.Title, &thumbnails, &post.Likes); err != nil {
			return posts, err
		}

		// Unserialize thumbnails
		var thumbs models.Thumbnails
		if err = json.Unmarshal(thumbnails, &thumbs); err != nil {
			return posts, fmt.Errorf("video ID '%s': %v", post.VideoID, err)
		}

		// Craft srcset string
		post.Srcset = thumbs.Srcset(480)
		post.Thumbnail = thumbs.Medium

		// Include the processed post in the result
		posts = append(posts, post)
	}

	// If error during iteration
	if err := rows.Err(); err != nil {
		return posts, err
	}

	return posts, err
}
