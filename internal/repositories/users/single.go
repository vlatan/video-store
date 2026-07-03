package users

import (
	"context"
	"time"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// Add or update a user
func (r *Repository) UpsertUser(ctx context.Context, u *models.User) (int, error) {

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
