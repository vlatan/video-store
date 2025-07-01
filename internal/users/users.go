package users

import (
	"context"
	"factual-docs/internal/services/database"
	"time"

	"github.com/markbates/goth"
)

type Service struct {
	db database.Service
}

func New(db database.Service) *Service {
	return &Service{db: db}
}

// Add or update a user
func (s *Service) UpsertUser(ctx context.Context, u *goth.User, analyticsID string) (int, error) {

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
	err := s.db.QueryRow(
		ctx,
		upsertUserQuery,
		NullString(&googleID),
		NullString(&facebookID),
		NullString(&analyticsID),
		NullString(&u.FirstName),
		NullString(&u.Email),
		NullString(&u.AvatarURL),
	).Scan(&id)

	return id, err
}

func (s *Service) DeleteUser(ctx context.Context, userID int) (int64, error) {
	return s.db.Exec(ctx, deleteUserQuery, userID)
}

func (s *Service) UpdateLastUserSeen(ctx context.Context, userID int, now time.Time) (int64, error) {
	return s.db.Exec(ctx, updateLastUserSeenQuery, userID, now)
}
