package database

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
func (s *service) GetUserActions(userID, postID int) (actions Actions, err error) {
	err = s.db.QueryRow(userActionsQuery, userID, postID).Scan(&actions)
	return actions, err
}

// func (s *service) Fave(userID, postID string) error
// func (s *service) Unfave(userID, postID string) error
// func (s *service) Like(userID, postID string) error
// func (s *service) Unlike(userID, postID string) error
// func (s *service) Edit(postID, title, desc string) error
// func (s *service) Delete(userID, postID string) error
