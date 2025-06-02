package database

type User struct {
	GoogleID    string
	FacebookID  string
	AnalyticsID string
	Name        string
	Email       string
	Picture     string
}

const addUserQuery = `
WITH existing_user AS (
    SELECT id FROM user 
    WHERE (google_id = $1 AND $1 IS NOT NULL) 
	OR (facebook_id = $2 AND $2 IS NOT NULL)
    OR (email = $5 AND AND $5 IS NOT NULL)
),
add_user AS (
    INSERT INTO users (google_id, facebook_id, analytics_id, name, email, picture)
    SELECT $1, $2, $3, $4, $5, $6
    WHERE NOT EXISTS (SELECT 1 FROM existing_user)
    RETURNING id
)
UPDATE user SET 
    google_id = COALESCE($1, google_id),
    facebook_id = COALESCE($2, facebook_id),
    analytics_id = COALESCE($3, analytics_id)
    name = $4,
    email = $5,
    picture = $6
    updated_at = NOW()
FROM existing_user
WHERE user.id = existing_user.id
RETURNING *
`

// Add or update a user
func (s *service) AddUser(u *User) (int64, error) {

	result, err := s.db.Exec(addUserQuery,
		u.GoogleID,
		u.FacebookID,
		u.AnalyticsID,
		u.Name,
		u.Email,
		u.Picture,
	)

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}
