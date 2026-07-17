package posts

import (
	"context"
	"database/sql"
	"errors"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// Get random posts, but exclude posts with the exact title match
func (r *Repository) GetRandomPosts(ctx context.Context, title string, limit int) (models.Posts, error) {

	var zero, posts models.Posts
	if title == "" {
		return zero, errors.New("title can't be empty string")
	}

	query, err := r.GetQuery("random_posts.sql", nil)
	if err != nil {
		return zero, err
	}

	// Get rows from DB
	rows, err := r.db.Pool.Query(ctx, query, title, limit)
	if err != nil {
		return zero, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	for rows.Next() {

		var (
			post          models.Post
			originalTitle sql.NullString
			avgRating     sql.NullFloat64
			ratingCount   sql.NullInt64
		)

		if err = rows.Scan(
			&post.VideoID,
			&post.Title,
			&originalTitle,
			&post.RawThumbs,
			&post.Likes,
			&avgRating,
			&ratingCount,
		); err != nil {
			return zero, err
		}

		// Include the processed post in the result
		post.OriginalTitle = utils.FromNullString(originalTitle)

		// Attach ratings if any
		if avgRating.Valid && ratingCount.Valid {
			post.Rating = &models.Rating{
				Avg:   utils.FromNullFloat64(avgRating),
				Count: utils.FromNullInt64(ratingCount),
			}
		}

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
