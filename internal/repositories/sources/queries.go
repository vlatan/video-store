package sources

const getSourcesQuery = `
	SELECT playlist_id, title, channel_title, channel_thumbnails FROM playlist
	WHERE id IN (SELECT DISTINCT playlist_db_id FROM post)
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
