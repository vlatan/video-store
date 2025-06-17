package database

import (
	"context"
)

type Actions struct {
	Liked bool
	Faved bool
}

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
func (s *service) GetUserActions(ctx context.Context, userID, postID int) (actions Actions, err error) {
	err = s.db.QueryRow(ctx, userActionsQuery, userID, postID).Scan(&actions.Liked, &actions.Faved)
	return actions, err
}

const likeQuery = `
	INSERT INTO post_like (user_id, post_id)
	SELECT $1, p.id 
	FROM post AS p 
	WHERE p.video_id = $2
`

// User likes a post
func (s *service) Like(ctx context.Context, userID int, videoID string) (int64, error) {
	result, err := s.db.Exec(ctx, likeQuery, userID, videoID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const unlikeQuery = `
	DELETE FROM post_like 
	USING post AS p 
	WHERE post_like.post_id = p.id 
	AND post_like.user_id = $1 
	AND p.video_id = $2
`

// User unlikes a post
func (s *service) Unlike(ctx context.Context, userID int, videoID string) (int64, error) {
	result, err := s.db.Exec(ctx, unlikeQuery, userID, videoID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const faveQuery = `
	INSERT INTO post_fave (user_id, post_id)
	SELECT $1, p.id 
	FROM post AS p 
	WHERE p.video_id = $2
`

// User favorites a post
func (s *service) Fave(ctx context.Context, userID int, videoID string) (int64, error) {
	result, err := s.db.Exec(ctx, faveQuery, userID, videoID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const unfaveQuery = `
	DELETE FROM post_fave 
	USING post AS p 
	WHERE post_fave.post_id = p.id 
	AND post_fave.user_id = $1 
	AND p.video_id = $2
`

// User unfavorites a post
func (s *service) Unfave(ctx context.Context, userID int, videoID string) (int64, error) {
	result, err := s.db.Exec(ctx, unfaveQuery, userID, videoID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// func (s *service) Edit(postID, title, desc string) error
// func (s *service) Delete(userID, postID string) error
