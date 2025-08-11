package users

import (
	"context"
	"factual-docs/internal/config"
	"factual-docs/internal/models"
	"factual-docs/internal/shared/database"
	"factual-docs/internal/utils"
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
		var googleID *string
		var facebookID *string
		var name *string
		var email *string
		var avatarURL *string
		var analytics_id *string
		var totalNum int

		// Get user row data to destination
		if err = rows.Scan(
			&googleID,
			&facebookID,
			&name,
			&email,
			&avatarURL,
			&analytics_id,
			&user.LastSeen,
			&user.CreatedAt,
			&totalNum,
		); err != nil {
			return nil, err
		}

		// Set user provider and user provider ID
		user.Provider = "google"
		user.UserID = utils.PtrToString(googleID)
		if fbID := utils.PtrToString(facebookID); fbID != "" {
			user.Provider = "facebook"
			user.UserID = fbID
		}

		// Convert the pointers back to strings
		user.Name = utils.PtrToString(name)
		user.Email = utils.PtrToString(email)
		user.AvatarURL = utils.PtrToString(avatarURL)

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
