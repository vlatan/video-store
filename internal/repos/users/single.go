package users

import (
	"context"
	"database/sql"

	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/utils/sqlnull"
)

// GetSingleUser gets single user based on ID
func (r *Repository) GetSingleUser(ctx context.Context, userID int) (types.User, error) {

	var zero, user types.User
	query, err := r.GetQuery("single_user.sql", nil)
	if err != nil {
		return zero, err
	}

	var name, email, avatarURL, publicID sql.NullString

	// Get user row data to destination
	if err := r.db.Pool.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.ProviderUserId,
		&user.Provider,
		&name,
		&email,
		&avatarURL,
		&publicID,
		&user.LastSeen,
		&user.CreatedAt,
	); err != nil {
		return zero, err
	}

	// Convert the NullString back to string
	user.Name = name.String
	user.Email = email.String
	user.AvatarURL = avatarURL.String
	user.PublicID = publicID.String

	return user, nil
}

// Add or update a user
func (r *Repository) UpsertUser(ctx context.Context, u *types.User) (int, error) {

	query, err := r.GetQuery("upsert_user.sql", nil)
	if err != nil {
		return 0, err
	}

	var id int
	err = r.db.Pool.QueryRow(
		ctx,
		query,
		u.ProviderUserId,
		u.Provider,
		u.PublicID,
		sqlnull.String(u.Name),
		sqlnull.String(u.Email),
		sqlnull.String(u.AvatarURL),
	).Scan(&id)

	return id, err
}

func (r *Repository) DeleteUser(ctx context.Context, userID int) (int64, error) {
	const query = "DELETE FROM app_user WHERE id = $1;"
	result, err := r.db.Pool.Exec(ctx, query, userID)
	return result.RowsAffected(), err
}

func (r *Repository) UpdateLastSeen(ctx context.Context, userID int) (int64, error) {
	const query = `
		UPDATE app_user 
		SET last_seen = now() 
		WHERE id = $1 AND last_seen < now() - interval '1 day'
	`
	result, err := r.db.Pool.Exec(ctx, query, userID)
	return result.RowsAffected(), err
}
