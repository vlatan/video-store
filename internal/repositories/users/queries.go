package users

const upsertUserQuery = `
	INSERT INTO app_user (
		provider_user_id, 
		provider, 
		analytics_id, 
		name, 
		email, 
		picture, 
		last_seen
	)
	VALUES ( $1, $2, $3, $4, $5, $6, NOW() ) 
	ON CONFLICT (provider, provider_user_id) 
	DO UPDATE SET 
		name = EXCLUDED.name,
		email = EXCLUDED.email,
		picture = EXCLUDED.picture,
		last_seen = NOW()
	RETURNING id;
`

const deleteUserQuery = "DELETE FROM app_user WHERE id = $1"

const updateLastUserSeenQuery = "UPDATE app_user SET last_seen = $2 WHERE id = $1"

const getUsersQuery = `
	SELECT
		provider_user_id,
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
