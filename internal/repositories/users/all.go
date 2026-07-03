package users

import (
	"context"
	"database/sql"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// Get users with limit and offset
func (r *Repository) GetUsers(ctx context.Context, page int) (*models.Users, error) {

	// Calculate the limit and offset
	limit := r.config.PostsPerPage
	offset := (page - 1) * limit

	query, err := r.GetQuery("offset_users.sql", nil)
	if err != nil {
		return nil, err
	}

	// Get rows from DB
	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	var users models.Users
	for rows.Next() {

		var totalNum int
		var user models.User
		var name, email, avatarURL, analyticsID sql.NullString

		// Get user row data to destination
		if err = rows.Scan(
			&user.ProviderUserId,
			&user.Provider,
			&name,
			&email,
			&avatarURL,
			&analyticsID,
			&user.LastSeen,
			&user.CreatedAt,
			&totalNum,
		); err != nil {
			return nil, err
		}

		// Convert the NullString back to string
		user.Name = utils.FromNullString(name)
		user.Email = utils.FromNullString(email)
		user.AvatarURL = utils.FromNullString(avatarURL)
		user.AnalyticsID = utils.FromNullString(analyticsID)

		// Include the user in the result
		users.Items = append(users.Items, user)
		if totalNum != 0 {
			users.TotalNum = totalNum
		}
	}

	// If error during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &users, nil
}
