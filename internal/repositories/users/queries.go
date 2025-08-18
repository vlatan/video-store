package users

const upsertUserQuery = `
	WITH existing_user AS (
		SELECT id FROM app_user 
		WHERE auth_id = $1
	),
	inserted AS (
		INSERT INTO app_user (
			auth_id, 
			analytics_id, 
			name, 
			email, 
			picture
		)
		SELECT $1, $2, $3, $4, $5
		WHERE NOT EXISTS (SELECT 1 FROM existing_user)
		RETURNING id
	),
	updated AS (
		UPDATE app_user SET 
			auth_id = COALESCE($1, auth_id),
			analytics_id = COALESCE($2, analytics_id),
			name = $3,
			email = $4,
			picture = $5,
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
		auth_id,
		provider, 
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
