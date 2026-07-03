package users

import (
	"context"

	"github.com/vlatan/video-store/internal/models"
)

// Check if the user liked and/or faved a post
func (r *Repository) GetUserActions(ctx context.Context, userID, postID int) (*models.Actions, error) {

	query, err := r.GetQuery("actions_user.sql", nil)
	if err != nil {
		return nil, err
	}

	var actions models.Actions
	row := r.db.Pool.QueryRow(ctx, query, userID, postID)
	err = row.Scan(&actions.Liked, &actions.Faved)
	return &actions, err
}
