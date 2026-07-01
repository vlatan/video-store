package posts

import (
	"context"
	"database/sql"
	"errors"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/repositories/queries"
	"github.com/vlatan/video-store/internal/utils"
)

// Get random posts, but exclude posts with the exact title match
func (r *Repository) GetRandomPosts(ctx context.Context, title string, limit int) (*models.Posts, error) {

	query, err := queries.Posts.Get("random_posts.sql", nil)
	if err != nil {
		return nil, err
	}

	if title == "" {
		return nil, errors.New("title can't be empty string")
	}

	// Get rows from DB
	rows, err := r.db.Pool.Query(ctx, query, title, limit)
	if err != nil {
		return nil, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	var posts models.Posts
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
			return nil, err
		}

		// Include the processed post in the result
		post.OriginalTitle = utils.FromNullString(originalTitle)
		posts.Items = append(posts.Items, post)
	}

	// If error during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Post-process the posts, prepare the thumbnail
	if err = postProcessPosts(ctx, posts); err != nil {
		return nil, err
	}

	return &posts, nil
}
