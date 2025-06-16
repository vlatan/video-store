package database

import (
	"database/sql"
	"time"

	"github.com/markbates/goth"
)

const upsertUserQuery = `
WITH existing_user AS (
    SELECT id FROM app_user 
    WHERE (google_id = $1 AND $1 IS NOT NULL) 
		OR (facebook_id = $2 AND $2 IS NOT NULL) 
		OR (email = $5 AND $5 IS NOT NULL)
),
inserted AS (
    INSERT INTO app_user (
        google_id, 
        facebook_id, 
        analytics_id, 
        name, 
        email, 
        picture
    )
    SELECT $1, $2, $3, $4, $5, $6
    WHERE NOT EXISTS (SELECT 1 FROM existing_user)
	RETURNING id
),
updated AS (
	UPDATE app_user SET 
		google_id = COALESCE($1, google_id),
		facebook_id = COALESCE($2, facebook_id),
		analytics_id = COALESCE($3, analytics_id),
		name = $4,
		email = $5,
		picture = $6,
        updated_at = CASE 
            WHEN COALESCE($1, google_id) IS DISTINCT FROM google_id 
				OR COALESCE($2, facebook_id) IS DISTINCT FROM facebook_id 
				OR COALESCE($3, analytics_id) IS DISTINCT FROM analytics_id 
				OR $4 IS DISTINCT FROM name 
				OR $5 IS DISTINCT FROM email 
				OR $6 IS DISTINCT FROM picture 
            THEN NOW() 
            ELSE updated_at 
        END,
		last_seen = NOW()
	FROM existing_user
	WHERE app_user.id = existing_user.id
	RETURNING app_user.id
)
SELECT id FROM inserted
UNION ALL
SELECT id FROM updated
`

// Add or update a user
func (s *service) UpsertUser(u *goth.User, analyticsID string) (int, error) {

	var (
		googleID   string
		facebookID string
	)

	switch u.Provider {
	case "google":
		googleID = u.UserID
	case "facebook":
		facebookID = u.UserID
	}

	var id int
	err := s.db.QueryRow(upsertUserQuery,
		NullString(&googleID),
		NullString(&facebookID),
		NullString(&analyticsID),
		NullString(&u.FirstName),
		NullString(&u.Email),
		NullString(&u.AvatarURL),
	).Scan(&id)

	return id, err
}

const updateLastUserSeenQuery = `
UPDATE app_user
SET last_seen = $2 WHERE id = $1
`

// Update the last seen column on a user
func (s *service) UpdateUserLastSeen(id int, t time.Time) error {
	_, err := s.db.Exec(updateLastUserSeenQuery, id, t)
	return err
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

type Actions struct {
	Liked bool
	Faved bool
}

// Check if the user liked a post
func (s *service) UserActions(userID, postID int) (actions Actions, err error) {
	err = s.db.QueryRow(userActionsQuery, userID, postID).Scan(&actions)
	return actions, err
}

// Helper function to convert string pointer or empty string to sql.NullString
func NullString(s *string) sql.NullString {
	if s == nil || *s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}
