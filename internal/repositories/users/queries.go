package users

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

const deleteUserQuery = "DELETE FROM app_user WHERE id = $1"

const updateLastUserSeenQuery = "UPDATE app_user SET last_seen = $2 WHERE id = $1"

const getUsersQuery = `
	SELECT
		google_id, 
		facebook_id, 
		name,
		email,
		picture,
		analytics_id,
		last_seen,
		created_at,
		COUNT(*) OVER() as total_results
	FROM app_user
	ORDER BY created_at
	LIMIT $1 OFFSET $2
`
