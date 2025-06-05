package database

import (
	"database/sql"

	"github.com/markbates/goth"
)

const upsertUserQuery = `
WITH existing_user AS (
    SELECT id FROM "user" 
    WHERE (google_id = $1 AND $1 IS NOT NULL) 
       OR (facebook_id = $2 AND $2 IS NOT NULL)
       OR (email = $5 AND $5 IS NOT NULL)
),
inserted AS (
    INSERT INTO "user" (
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
	UPDATE "user" SET 
		google_id = COALESCE($1, google_id),
		facebook_id = COALESCE($2, facebook_id),
		analytics_id = COALESCE($3, analytics_id),
		name = $4,
		email = $5,
		picture = $6,
		updated_at = NOW(),
		last_seen = NOW()
	FROM existing_user
	WHERE "user".id = existing_user.id
	RETURNING "user".id
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

// Helper function to convert string pointer or empty string to sql.NullString
func NullString(s *string) sql.NullString {
	if s == nil || *s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

func (s *service) UpdateUserLastSeen(id int) error {
	_, err := s.db.Exec("UPDATE 'user' SET last_seen = NOW() WHERE id = $1", id)
	return err
}
