package repositories

import (
	"context"
	"factual-docs/internal/shared/database"
	"time"

	"github.com/markbates/goth"
)

type User struct {
	db database.Service
}

func NewUserRepo(db database.Service) *User {
	return &User{db: db}
}

// Add or update a user
func (u *User) UpsertUser(ctx context.Context, gu *goth.User, analyticsID string) (int, error) {

	var (
		googleID   string
		facebookID string
	)

	switch gu.Provider {
	case "google":
		googleID = gu.UserID
	case "facebook":
		facebookID = gu.UserID
	}

	var id int
	err := u.db.QueryRow(
		ctx,
		upsertUserQuery,
		NullString(&googleID),
		NullString(&facebookID),
		NullString(&analyticsID),
		NullString(&gu.FirstName),
		NullString(&gu.Email),
		NullString(&gu.AvatarURL),
	).Scan(&id)

	return id, err
}

func (u *User) DeleteUser(ctx context.Context, userID int) (int64, error) {
	return u.db.Exec(ctx, deleteUserQuery, userID)
}

func (u *User) UpdateLastUserSeen(ctx context.Context, userID int, now time.Time) (int64, error) {
	return u.db.Exec(ctx, updateLastUserSeenQuery, userID, now)
}
