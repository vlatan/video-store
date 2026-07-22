package posts

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// User likes a post
func (r *Repository) Like(ctx context.Context, userID int, videoID string) (int64, error) {

	query, err := r.GetQuery("like_post.sql", nil)
	if err != nil {
		return 0, err
	}

	result, err := r.db.Pool.Exec(ctx, query, userID, videoID)
	return result.RowsAffected(), err
}

// User unlikes a post
func (r *Repository) Unlike(ctx context.Context, userID int, videoID string) (int64, error) {

	query, err := r.GetQuery("unlike_post.sql", nil)
	if err != nil {
		return 0, err
	}

	result, err := r.db.Pool.Exec(ctx, query, userID, videoID)
	return result.RowsAffected(), err
}

// User favorites a post
func (r *Repository) Fave(ctx context.Context, userID int, videoID string) (int64, error) {

	query, err := r.GetQuery("fave_post.sql", nil)
	if err != nil {
		return 0, err
	}

	result, err := r.db.Pool.Exec(ctx, query, userID, videoID)
	return result.RowsAffected(), err
}

// User unfavorites a post
func (r *Repository) Unfave(ctx context.Context, userID int, videoID string) (int64, error) {

	query, err := r.GetQuery("unfave_post.sql", nil)
	if err != nil {
		return 0, err
	}

	result, err := r.db.Pool.Exec(ctx, query, userID, videoID)
	return result.RowsAffected(), err
}

// Rate records user's post rating
func (r *Repository) Rate(ctx context.Context, rating uint8, userID int, videoID string) (models.Rating, error) {

	var zero, rd models.Rating

	// Start trannsaction
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return zero, err
	}

	// Rollback if something goes wrong.
	// Release the connection in any case.
	defer func() {
		rbErr := tx.Rollback(ctx)
		if rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			slog.ErrorContext(
				ctx, "transaction rollback on post rating failed",
				"userId", userID,
				"postId", videoID,
				"error", rbErr,
			)
		}
	}()

	query, err := r.GetQuery("rate_post.sql", nil)
	if err != nil {
		return zero, err
	}

	var postId int64
	err = tx.QueryRow(ctx, query, rating, userID, videoID).Scan(&postId)
	if err != nil {
		return zero, err
	}

	query = `
		SELECT ROUND(AVG(rating), 2)::float8, COUNT(*)
		FROM post_rating WHERE post_id = $1
	`
	err = tx.QueryRow(ctx, query, postId).Scan(&rd.Avg, &rd.Count)
	if err != nil {
		return zero, err
	}

	// Commit the changes
	if err := tx.Commit(ctx); err != nil {
		return zero, err
	}

	return rd, nil
}

// Update a playlist
func (r *Repository) UpdateSource(ctx context.Context, videoID, playlistID string) (int64, error) {

	query, err := r.GetQuery("update_post_source.sql", nil)
	if err != nil {
		return 0, err
	}

	result, err := r.db.Pool.Exec(ctx, query, videoID, playlistID)
	return result.RowsAffected(), err
}

// Update post description
func (r *Repository) UpdateGeneratedData(ctx context.Context, post *models.Post) (int64, error) {

	query, err := r.GetQuery("update_post.sql", nil)
	if err != nil {
		return 0, err
	}

	result, err := r.db.Pool.Exec(
		ctx,
		query,
		post.VideoID,
		utils.ToNullString(post.OriginalTitle),
		post.Category.Name,
		post.Summary,
	)

	return result.RowsAffected(), err
}

// Ban a post (move it to deleted table)
func (r *Repository) BanPost(ctx context.Context, videoID string) (int64, error) {

	query, err := r.GetQuery("ban_post.sql", nil)
	if err != nil {
		return 0, err
	}

	result, err := r.db.Pool.Exec(ctx, query, videoID)
	return result.RowsAffected(), err
}

// Delete a post
func (r *Repository) DeletePost(ctx context.Context, videoID string) (int64, error) {
	const query = "DELETE FROM post WHERE video_id = $1;"
	result, err := r.db.Pool.Exec(ctx, query, videoID)
	return result.RowsAffected(), err
}
