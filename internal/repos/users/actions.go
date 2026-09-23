package users

import (
	"context"
	"database/sql"

	"github.com/vlatan/video-store/internal/types"
)

// Check if the user liked and/or faved a post
func (r *Repository) GetUserActions(ctx context.Context, userID, postID int) (types.Actions, error) {

	var zero, actions types.Actions
	query, err := r.GetQuery("actions_user.sql", nil)
	if err != nil {
		return zero, err
	}

	var whenFaved sql.NullTime
	row := r.db.Pool.QueryRow(ctx, query, userID, postID)
	err = row.Scan(
		&actions.UserID,
		&actions.PostID,
		&actions.Liked,
		&actions.Faved,
		&whenFaved,
		&actions.Review.Rating,
		&actions.Review.Headline,
		&actions.Review.Content,
	)

	if err != nil {
		return zero, err
	}

	actions.WhenFaved = whenFaved.Time
	return actions, nil
}
