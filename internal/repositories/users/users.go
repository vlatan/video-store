package users

import (
	"context"
	"factual-docs/internal/shared/database"
	"factual-docs/internal/shared/utils"
	"time"

	"github.com/markbates/goth"
)

type Repository struct {
	db database.Service
}

func New(db database.Service) *Repository {
	return &Repository{db: db}
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
