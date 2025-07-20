package sources

const sourceExistsQuery = `
	SELECT 1 FROM playlist
	WHERE playlist_id = $1
`

const getSourcesQuery = `
	SELECT
		playlist_id, 
		title, 
		channel_title, 
		channel_thumbnails,
		updated_at
	FROM playlist
	ORDER BY id DESC
`

const insertSourceQuery = `
	INSERT INTO playlist (
		playlist_id, 
		channel_id,
		title,
		channel_title,
		thumbnails,
		channel_thumbnails,
		description,
		channel_description,
		user_id
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULLIF($9, 0))
`
