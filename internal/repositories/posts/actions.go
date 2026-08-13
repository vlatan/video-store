package posts

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/microcosm-cc/bluemonday"
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

// Rate records user's post rating,
// Returns a struct with rating count and average rating for the video.
func (r *Repository) Rate(ctx context.Context, rating uint8, userID int, videoID string) (models.RatingStats, error) {

	var zero, rs models.RatingStats

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
	err = tx.QueryRow(ctx, query, postId).Scan(&rs.Avg, &rs.Count)
	if err != nil {
		return zero, err
	}

	// Commit the changes
	if err := tx.Commit(ctx); err != nil {
		return zero, err
	}

	return rs, nil
}

// DeleteRating deletes user rating and by cascade deletes their review too.
// Returns updated rating stats.
func (r *Repository) DeleteRating(
	ctx context.Context,
	userID int,
	videoID string) (models.RatingStats, error) {

	var zero, rs models.RatingStats

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
				ctx, "transaction rollback on delete user rating failed",
				"userId", userID,
				"postId", videoID,
				"error", rbErr,
			)
		}
	}()

	query, err := r.GetQuery("delete_rating.sql", nil)
	if err != nil {
		return zero, err
	}

	var postId int64
	err = tx.QueryRow(ctx, query, userID, videoID).Scan(&postId)
	if err != nil {
		return zero, err
	}

	query = `
		SELECT ROUND(AVG(rating), 2)::float8, COUNT(*)
		FROM post_rating WHERE post_id = $1
	`
	err = tx.QueryRow(ctx, query, postId).Scan(&rs.Avg, &rs.Count)
	if err != nil {
		return zero, err
	}

	// Commit the changes
	if err := tx.Commit(ctx); err != nil {
		return zero, err
	}

	return rs, nil
}

// Review records user's post rating and review.
// Returns a map with review data and ratings stats for the post.
func (r *Repository) Review(
	ctx context.Context,
	userID int, videoID string,
	rating uint8, headline, content string) (map[string]any, error) {

	var rs models.RatingStats
	var re models.Review

	// Start trannsaction
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	// Rollback if something goes wrong.
	// Release the connection in any case.
	defer func() {
		rbErr := tx.Rollback(ctx)
		if rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			slog.ErrorContext(
				ctx, "transaction rollback on post review failed",
				"userId", userID,
				"postId", videoID,
				"error", rbErr,
			)
		}
	}()

	query, err := r.GetQuery("review_post.sql", nil)
	if err != nil {
		return nil, err
	}

	// Insert review
	var postId int64
	err = tx.QueryRow(ctx, query, rating, userID, videoID, headline, content).Scan(&postId)
	if err != nil {
		return nil, err
	}

	query = `
		SELECT ROUND(AVG(rating), 2)::float8, COUNT(*)
		FROM post_rating WHERE post_id = $1
	`
	err = tx.QueryRow(ctx, query, postId).Scan(&rs.Avg, &rs.Count)
	if err != nil {
		return nil, err
	}

	// Commit the changes
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	// Define policies for review headline and content sanitization
	strictPolicy := bluemonday.StrictPolicy()
	simplePolicy := utils.SimplePolicy()

	// Sanitize headline
	safeHeadline := strictPolicy.Sanitize(headline)
	re.HTMLHeadline = template.HTML(safeHeadline) // #nosec G203

	// Convert to HTML and sanitize content
	safeHTMLcontent, err := utils.ParseMarkdown(content, simplePolicy)
	if err != nil {
		return nil, fmt.Errorf(
			"could not parse/sanitize markdown on video %q review: %v",
			videoID, err,
		)
	}

	re.HTMLContent = template.HTML(safeHTMLcontent) // #nosec G203

	result := map[string]any{
		"review": re,
		"stats":  rs,
	}

	return result, nil
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
