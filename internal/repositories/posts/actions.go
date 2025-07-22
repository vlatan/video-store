package posts

import (
	"context"
	"factual-docs/internal/models"
)

// Check if the user liked and/or faved a post
func (r *Repository) GetUserActions(ctx context.Context, userID, postID int) (actions models.Actions, err error) {
	err = r.db.QueryRow(ctx, userActionsQuery, userID, postID).Scan(&actions.Liked, &actions.Faved)
	return actions, err
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

// Delete a post
func (r *Repository) BanPost(ctx context.Context, videoID string) (int64, error) {
	return r.db.Exec(ctx, banPostQuery, videoID)
}
