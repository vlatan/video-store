package posts

import (
	"context"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

const likeQuery = `
	INSERT INTO post_like (user_id, post_id)
	SELECT $1, p.id 
	FROM post AS p 
	WHERE p.video_id = $2
`

// User likes a post
func (r *Repository) Like(ctx context.Context, userID int, videoID string) (int64, error) {
	result, err := r.db.Pool.Exec(ctx, likeQuery, userID, videoID)
	return result.RowsAffected(), err
}

const unlikeQuery = `
	DELETE FROM post_like 
	USING post AS p 
	WHERE post_like.post_id = p.id 
	AND post_like.user_id = $1 
	AND p.video_id = $2
`

// User unlikes a post
func (r *Repository) Unlike(ctx context.Context, userID int, videoID string) (int64, error) {
	result, err := r.db.Pool.Exec(ctx, unlikeQuery, userID, videoID)
	return result.RowsAffected(), err
}

const faveQuery = `
	INSERT INTO post_fave (user_id, post_id)
	SELECT $1, p.id 
	FROM post AS p 
	WHERE p.video_id = $2
`

// User favorites a post
func (r *Repository) Fave(ctx context.Context, userID int, videoID string) (int64, error) {
	result, err := r.db.Pool.Exec(ctx, faveQuery, userID, videoID)
	return result.RowsAffected(), err
}

const unfaveQuery = `
	DELETE FROM post_fave 
	USING post AS p 
	WHERE post_fave.post_id = p.id 
	AND post_fave.user_id = $1 
	AND p.video_id = $2
`

// User unfavorites a post
func (r *Repository) Unfave(ctx context.Context, userID int, videoID string) (int64, error) {
	result, err := r.db.Pool.Exec(ctx, unfaveQuery, userID, videoID)
	return result.RowsAffected(), err
}

const updatePlaylistQuery = `
	UPDATE post
	SET 
		playlist_id = $2,
		playlist_db_id = (
			SELECT id
			FROM playlist
			WHERE playlist_id = $2::VARCHAR(50)
		)
	WHERE video_id = $1
`

// Update a playlist
func (r *Repository) UpdateSource(ctx context.Context, videoID, playlistID string) (int64, error) {
	result, err := r.db.Pool.Exec(ctx, updatePlaylistQuery, videoID, playlistID)
	return result.RowsAffected(), err
}

const updateGeneratedDataQuery = `
	UPDATE post
	SET
		original_title = $2,
		category_id = (SELECT id FROM category WHERE name = $3),
		summary = $4
	WHERE video_id = $1
`

// Update post description
func (r *Repository) UpdateGeneratedData(ctx context.Context, post *models.Post) (int64, error) {

	result, err := r.db.Pool.Exec(
		ctx,
		updateGeneratedDataQuery,
		post.VideoID,
		utils.ToNullString(post.OriginalTitle),
		post.Category.Name,
		post.Summary,
	)

	return result.RowsAffected(), err
}

const banPostQuery = `
	WITH dp AS (
		DELETE FROM post
		WHERE video_id = $1
		RETURNING video_id, NULLIF(provider, '') as provider
	)
	INSERT INTO deleted_post (video_id, provider)
	SELECT video_id, provider FROM dp
`

// Ban a post (move it to deleted table)
func (r *Repository) BanPost(ctx context.Context, videoID string) (int64, error) {
	result, err := r.db.Pool.Exec(ctx, banPostQuery, videoID)
	return result.RowsAffected(), err
}

const deletePostQuery = `
	DELETE FROM post
	WHERE video_id = $1
`

// Delete a post
func (r *Repository) DeletePost(ctx context.Context, videoID string) (int64, error) {
	result, err := r.db.Pool.Exec(ctx, deletePostQuery, videoID)
	return result.RowsAffected(), err
}
