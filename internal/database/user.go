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
)
UPDATE "user" SET 
    google_id = COALESCE($1, google_id),
    facebook_id = COALESCE($2, facebook_id),
    analytics_id = COALESCE($3, analytics_id),
    name = $4,
    email = $5,
    picture = $6,
    updated_at = NOW()
FROM existing_user
WHERE "user".id = existing_user.id;
`

// Add or update a user
func (s *service) UpsertUser(u *goth.User, analyticsID string) (int64, error) {

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

	result, err := s.db.Exec(upsertUserQuery,
		NullString(&googleID),
		NullString(&facebookID),
		NullString(&analyticsID),
		NullString(&u.FirstName),
		NullString(&u.Email),
		NullString(&u.AvatarURL),
	)

	if err != nil {
		return 0, err
	}

	num, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return num, nil
}

// Helper function to convert string pointer or empty string to sql.NullString
func NullString(s *string) sql.NullString {
	if s == nil || *s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}
