package users

import (
	"context"

	"github.com/vlatan/video-store/internal/models"
)

// Check if the user liked and/or faved a post
func (r *Repository) GetUserActions(ctx context.Context, userID, postID int) (models.Actions, error) {

	var zero, actions models.Actions
	query, err := r.GetQuery("actions_user.sql", nil)
	if err != nil {
		return zero, err
	}

	row := r.db.Pool.QueryRow(ctx, query, userID, postID)
	err = row.Scan(
		&actions.UserID,
		&actions.PostID,
		&actions.Liked,
		&actions.Faved,
		&actions.WhenFaved,
		&actions.Rating,
		&actions.Review.Headline,
		&actions.Review.Content,
	)

	if err != nil {
		return zero, err
	}

	return actions, nil
}
