package users

import (
	"context"
	"database/sql"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// Check if the user liked and/or faved a post
func (r *Repository) GetUserActions(ctx context.Context, userID, postID int) (models.Actions, error) {

	var zero, actions models.Actions
	query, err := r.GetQuery("actions_user.sql", nil)
	if err != nil {
		return zero, err
	}

	var headline, content sql.NullString
	row := r.db.Pool.QueryRow(ctx, query, userID, postID)
	err = row.Scan(
		&actions.UserID,
		&actions.PostID,
		&actions.Liked,
		&actions.Faved,
		&actions.WhenFaved,
		&actions.Rating,
		&headline,
		&content,
	)

	if err != nil {
		return zero, err
	}

	if headline.Valid && content.Valid {
		actions.Review = &models.Review{
			Headline: utils.FromNullString(headline),
			Content:  utils.FromNullString(content),
		}
	}

	return actions, nil
}
