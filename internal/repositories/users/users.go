package users

import (
	"context"
	"database/sql"
	"time"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

type Repository struct {
	db     *database.Service
	config *config.Config
}

func New(db *database.Service, config *config.Config) *Repository {
	return &Repository{
		db:     db,
		config: config,
	}
}

// Add or update a user
func (r *Repository) UpsertUser(ctx context.Context, u *models.User) (int, error) {

	var id int
	err := r.db.Pool.QueryRow(
		ctx,
		upsertUserQuery,
		u.ProviderUserId,
		u.Provider,
		utils.ToNullString(u.AnalyticsID),
		utils.ToNullString(u.Name),
		utils.ToNullString(u.Email),
		utils.ToNullString(u.AvatarURL),
	).Scan(&id)

	return id, err
}

func (r *Repository) DeleteUser(ctx context.Context, userID int) (int64, error) {
	result, err := r.db.Pool.Exec(ctx, deleteUserQuery, userID)
	return result.RowsAffected(), err
}

func (r *Repository) UpdateLastUserSeen(ctx context.Context, userID int, now time.Time) (int64, error) {
	result, err := r.db.Pool.Exec(ctx, updateLastUserSeenQuery, userID, now)
	return result.RowsAffected(), err
}

// Get users with limit and offset
func (r *Repository) GetUsers(ctx context.Context, page int) (*models.Users, error) {

	var users models.Users

	// Calculate the limit and offset
	limit := r.config.PostsPerPage
	offset := (page - 1) * limit

	// Get rows from DB
	rows, err := r.db.Pool.Query(ctx, getUsersQuery, limit, offset)
	if err != nil {
		return nil, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
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
