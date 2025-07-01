package users

import (
	"context"
	"factual-docs/internal/services/database"

	"github.com/markbates/goth"
)

type Repository struct {
	DB database.Service
}

func NewRepository(db database.Service) *Repository {
	return &Repository{DB: db}
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
	err := r.DB.QueryRow(
		ctx,
		UpsertUserQuery,
		NullString(&googleID),
		NullString(&facebookID),
		NullString(&analyticsID),
		NullString(&u.FirstName),
		NullString(&u.Email),
		NullString(&u.AvatarURL),
	).Scan(&id)

	return id, err
}
