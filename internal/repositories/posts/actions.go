package posts

import (
	"context"

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
func (r *Repository) Rate(ctx context.Context, rating, userID int, videoID string) (int64, error) {

	query, err := r.GetQuery("rate_post.sql", nil)
	if err != nil {
		return 0, err
	}

	result, err := r.db.Pool.Exec(ctx, query, rating, userID, videoID)
	return result.RowsAffected(), err
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
