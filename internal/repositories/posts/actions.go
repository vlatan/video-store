package posts

import (
	"context"

	"github.com/vlatan/video-store/internal/models"
)

const userActionsQuery = `
	SELECT 
		EXISTS (
			SELECT 1 FROM post_like
			WHERE user_id = $1 AND post_id = $2
		) AS liked,
		EXISTS (
			SELECT 1 FROM post_fave
			WHERE user_id = $1 AND post_id = $2
		) AS faved
`

// Check if the user liked and/or faved a post
func (r *Repository) GetUserActions(ctx context.Context, userID, postID int) (*models.Actions, error) {
	row := r.db.Pool.QueryRow(ctx, userActionsQuery, userID, postID)
	var actions models.Actions
	err := row.Scan(&actions.Liked, &actions.Faved)
	return &actions, err
}

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

const updateTitleQuery = `
	UPDATE post
	SET title = $2
	WHERE video_id = $1
`

// Update post title
func (r *Repository) UpdateTitle(ctx context.Context, videoID, title string) (int64, error) {
	result, err := r.db.Pool.Exec(ctx, updateTitleQuery, videoID, title)
	return result.RowsAffected(), err
}

const updateDescQuery = `
	UPDATE post
	SET short_description = $2
	WHERE video_id = $1
`

// Update post description
func (r *Repository) UpdateDesc(ctx context.Context, videoID, description string) (int64, error) {
	result, err := r.db.Pool.Exec(ctx, updateDescQuery, videoID, description)
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
func (r *Repository) UpdatePlaylist(ctx context.Context, videoID, playlistID string) (int64, error) {
	result, err := r.db.Pool.Exec(ctx, updatePlaylistQuery, videoID, playlistID)
	return result.RowsAffected(), err
}

const updateGeneretedDataQuery = `
	UPDATE post
	SET
		category_id = (SELECT id FROM category WHERE name = $2),
		short_description = $3
	WHERE video_id = $1
`

// Update post description
func (r *Repository) UpdateGeneratedData(ctx context.Context, post *models.Post) (int64, error) {
	result, err := r.db.Pool.Exec(
		ctx,
		updateGeneretedDataQuery,
		post.VideoID,
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
