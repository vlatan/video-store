package users

import (
	"context"
	"factual-docs/internal/models"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/database"
	"factual-docs/internal/shared/utils"
	"time"

	"github.com/markbates/goth"
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
func (r *Repository) UpsertUser(ctx context.Context, u *goth.User, analyticsID string) (int, error) {

	var (
		googleID   string
		facebookID string
	)

	switch u.Provider {
	case "google":
		googleID = u.UserID
	case "facebook":
		facebookID = u.UserID
	}

	var id int
	err := r.db.QueryRow(
		ctx,
		upsertUserQuery,
		utils.NullString(&googleID),
		utils.NullString(&facebookID),
		utils.NullString(&analyticsID),
		utils.NullString(&u.FirstName),
		utils.NullString(&u.Email),
		utils.NullString(&u.AvatarURL),
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
func (r *Repository) GetUsers(ctx context.Context, page int) (users []models.User, err error) {

	// Calculate the limit and offset
	limit := r.config.PostsPerPage
	offset := (page - 1) * limit

	// Get rows from DB
	rows, err := r.db.Query(ctx, getUsersQuery, limit, offset)
	if err != nil {
		return users, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	for rows.Next() {
		var user models.User
		var googleID string
		var facebookID string

		// Get user row data to destination
		if err = rows.Scan(
			&googleID,
			&facebookID,
			&user.Name,
			&user.Email,
			&user.AvatarURL,
			user.LastSeen,
			user.CreatedAt,
		); err != nil {
			return users, err
		}

		// Include the processed post in the result
		users = append(users, user)
	}

	// If error during iteration
	if err = rows.Err(); err != nil {
		return users, err
	}

	return users, err
}
