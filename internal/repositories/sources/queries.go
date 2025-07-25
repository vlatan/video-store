package sources

const sourceExistsQuery = `
	SELECT 1 FROM playlist
	WHERE playlist_id = $1
`

const getSourcesQuery = `
	SELECT
		playlist_id,
		channel_id,
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

const updateSourceQuery = `
	UPDATE playlist
	SET
		channel_id = $2,
		title = $3,
		channel_title = $4,
		thumbnails = $5,
		channel_thumbnails = $6,
		description = $7,
		channel_description = $8,
		updated_at = NOW()
	WHERE playlist_id = $1
`
