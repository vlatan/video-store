package database

import "context"

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
	err = s.db.QueryRow(ctx, userActionsQuery, userID, postID).Scan(&actions)
	return actions, err
}

const likeQuery = `
	INSERT INTO post_like (user_id, post_id)
	SELECT $1, p.id 
	FROM post AS p 
	WHERE p.video_id = $2
	RETURNING *
`

func (s *service) Like(ctx context.Context, userID int, videoID string) (int64, error) {
	result, err := s.db.Exec(ctx, likeQuery, userID, videoID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// func (s *service) Unlike(userID, postID string) error
// func (s *service) Fave(userID, postID string) error
// func (s *service) Unfave(userID, postID string) error
// func (s *service) Edit(postID, title, desc string) error
// func (s *service) Delete(userID, postID string) error
