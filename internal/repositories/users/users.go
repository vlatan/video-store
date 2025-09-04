package users

import (
	"context"
	"factual-docs/internal/config"
	"factual-docs/internal/drivers/database"
	"factual-docs/internal/models"
	"factual-docs/internal/utils"
	"time"
)

type Repository struct {
	db     database.Service
	config *config.Config
}

func New(db database.Service, config *config.Config) *Repository {
	return &Repository{
		db:     db,
		config: config,
	}
}

// Add or update a user
func (r *Repository) UpsertUser(ctx context.Context, u *models.User) (int, error) {

	var id int
	err := r.db.QueryRow(
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
	return r.db.Exec(ctx, deleteUserQuery, userID)
}

func (r *Repository) UpdateLastUserSeen(ctx context.Context, userID int, now time.Time) (int64, error) {
	return r.db.Exec(ctx, updateLastUserSeenQuery, userID, now)
}

// Get users with limit and offset
func (r *Repository) GetUsers(ctx context.Context, page int) (*models.Users, error) {

	var users models.Users

	// Calculate the limit and offset
	limit := r.config.PostsPerPage
	offset := (page - 1) * limit

	// Get rows from DB
	rows, err := r.db.Query(ctx, getUsersQuery, limit, offset)
	if err != nil {
		return nil, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	for rows.Next() {

		var user models.User

		// All these are nullable in the DB, so we
		// temp use pointers to accept NULL values
		var name *string
		var email *string
		var avatarURL *string
		var analyticsID *string
		var totalNum int

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

		// Convert the pointers back to strings
		user.Name = utils.PtrToString(name)
		user.Email = utils.PtrToString(email)
		user.AvatarURL = utils.PtrToString(avatarURL)
		user.AnalyticsID = utils.PtrToString(analyticsID)

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
