package users

import (
	"context"
	"database/sql"
	"time"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/repositories/queries"
	"github.com/vlatan/video-store/internal/utils"
)

type Repository struct {
	db     *database.Service
	config *config.Config
}

func New(db *database.Service, config *config.Config) *Repository {
	return &Repository{db, config}
}

// Add or update a user
func (r *Repository) UpsertUser(ctx context.Context, u *models.User) (int, error) {

	query, err := queries.GetQuery("upsert_user.sql", nil)
	if err != nil {
		return 0, err
	}

	var id int
	err = r.db.Pool.QueryRow(
		ctx,
		query,
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
	const query = "DELETE FROM app_user WHERE id = $1;"
	result, err := r.db.Pool.Exec(ctx, query, userID)
	return result.RowsAffected(), err
}

func (r *Repository) UpdateLastUserSeen(ctx context.Context, userID int, now time.Time) (int64, error) {
	const query = "UPDATE app_user SET last_seen = $2 WHERE id = $1"
	result, err := r.db.Pool.Exec(ctx, query, userID, now)
	return result.RowsAffected(), err
}

// Get users with limit and offset
func (r *Repository) GetUsers(ctx context.Context, page int) (*models.Users, error) {

	// Calculate the limit and offset
	limit := r.config.PostsPerPage
	offset := (page - 1) * limit

	query, err := queries.GetQuery("offset_users.sql", nil)
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

// Check if the user liked and/or faved a post
func (r *Repository) GetUserActions(ctx context.Context, userID, postID int) (*models.Actions, error) {

	query, err := queries.GetQuery("actions_user.sql", nil)
	if err != nil {
		return nil, err
	}

	var actions models.Actions
	row := r.db.Pool.QueryRow(ctx, query, userID, postID)
	err = row.Scan(&actions.Liked, &actions.Faved)
	return &actions, err
}
