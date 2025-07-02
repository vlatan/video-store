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

const LikeQuery = `
	INSERT INTO post_like (user_id, post_id)
	SELECT $1, p.id 
	FROM post AS p 
	WHERE p.video_id = $2
`

const UnlikeQuery = `
	DELETE FROM post_like 
	USING post AS p 
	WHERE post_like.post_id = p.id 
	AND post_like.user_id = $1 
	AND p.video_id = $2
`

const FaveQuery = `
	INSERT INTO post_fave (user_id, post_id)
	SELECT $1, p.id 
	FROM post AS p 
	WHERE p.video_id = $2
`

const UnfaveQuery = `
	DELETE FROM post_fave 
	USING post AS p 
	WHERE post_fave.post_id = p.id 
	AND post_fave.user_id = $1 
	AND p.video_id = $2
`

const UpdateTitleQuery = `
	UPDATE post
	SET title = $2, updated_at = NOW()
	WHERE video_id = $1
`

const UpdateDescQuery = `
	UPDATE post
	SET short_description = $2, updated_at = NOW()
	WHERE video_id = $1
`

const DeletePostQuery = `
	WITH dp AS (
		DELETE FROM post
		WHERE video_id = $1
		RETURNING video_id
	)
	INSERT INTO deleted_post (video_id)
	SELECT video_id FROM dp;
`
