package posts

import (
	"context"

	"github.com/vlatan/video-store/internal/models"
)

// GetPostReviews gets posts reviews
func (r *Repository) GetPostReviews(ctx context.Context, videoID string) ([]models.Review, error) {

	query, err := r.GetQuery("post_reviews.sql", nil)
	if err != nil {
		return nil, err
	}

	// Get rows from DB
	var reviews []models.Review
	rows, err := r.db.Pool.Query(ctx, query, videoID)
	if err != nil {
		return nil, err
	}

	return reviews, nil
}
