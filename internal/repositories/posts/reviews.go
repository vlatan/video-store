package posts

import (
	"context"
	"database/sql"

	"github.com/vlatan/video-store/internal/models"
)

// GetPostReviews gets posts reviews
func (r *Repository) GetPostReviews(ctx context.Context, videoID string) (models.Reviews, error) {

	query, err := r.GetQuery("post_reviews.sql", nil)
	if err != nil {
		return nil, err
	}

	// Get rows from DB
	var reviews models.Reviews
	rows, err := r.db.Pool.Query(ctx, query, videoID)
	if err != nil {
		return nil, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	for rows.Next() {

		var (
			review models.Review
			name,
			email,
			avatarURL,
			analyticsID sql.NullString
		)

		err = rows.Scan(
			&review.User.ID,
			&review.User.ProviderUserId,
			&review.User.Provider,
			&name,
			&email,
			&avatarURL,
			&analyticsID,
			&review.Headline,
			&review.Content,
			&review.Rating,
			&review.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		review.User.Name = name.String
		review.User.Email = email.String
		review.User.AvatarURL = avatarURL.String
		review.User.AnalyticsID = analyticsID.String
		reviews = append(reviews, review)
	}

	// If error during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return reviews, nil
}
