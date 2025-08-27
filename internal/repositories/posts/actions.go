package posts

import (
	"context"
	"factual-docs/internal/models"
)

// Check if the user liked and/or faved a post
func (r *Repository) GetUserActions(ctx context.Context, userID, postID int) (*models.Actions, error) {
	row := r.db.QueryRow(ctx, userActionsQuery, userID, postID)
	var actions models.Actions
	err := row.Scan(&actions.Liked, &actions.Faved)
	return &actions, err
}

// User likes a post
func (r *Repository) Like(ctx context.Context, userID int, videoID string) (int64, error) {
	return r.db.Exec(ctx, likeQuery, userID, videoID)
}

// User unlikes a post
func (r *Repository) Unlike(ctx context.Context, userID int, videoID string) (int64, error) {
	return r.db.Exec(ctx, unlikeQuery, userID, videoID)
}

// User favorites a post
func (r *Repository) Fave(ctx context.Context, userID int, videoID string) (int64, error) {
	return r.db.Exec(ctx, faveQuery, userID, videoID)
}

// User unfavorites a post
func (r *Repository) Unfave(ctx context.Context, userID int, videoID string) (int64, error) {
	return r.db.Exec(ctx, unfaveQuery, userID, videoID)
}

// Update post title
func (r *Repository) UpdateTitle(ctx context.Context, videoID, title string) (int64, error) {
	return r.db.Exec(ctx, updateTitleQuery, videoID, title)
}

// Update post description
func (r *Repository) UpdateDesc(ctx context.Context, videoID, description string) (int64, error) {
	return r.db.Exec(ctx, updateDescQuery, videoID, description)
}

// Update a playlist
func (r *Repository) UpdatePlaylist(ctx context.Context, videoID, playlistID string) (int64, error) {
	return r.db.Exec(ctx, updatePlaylistQuery, videoID, playlistID)
}

// Update post description
func (r *Repository) UpdateGeneratedData(ctx context.Context, post *models.Post) (int64, error) {
	return r.db.Exec(ctx, updateGeneretedDataQuery, post.VideoID, post.Category.Name, post.ShortDesc)
}

// Delete a post
func (r *Repository) BanPost(ctx context.Context, videoID string) (int64, error) {
	return r.db.Exec(ctx, banPostQuery, videoID)
}

// Ban a post (move it to deleted table)
func (r *Repository) DeletePost(ctx context.Context, videoID string) (int64, error) {
	return r.db.Exec(ctx, deletePostQuery, videoID)
}
