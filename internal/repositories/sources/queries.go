package sources

const postExistsQuery = `
	SELECT 1 FROM playlist
	WHERE playlist_id = $1
`

const getSourcesQuery = `
	SELECT
		p.playlist_id, 
		p.title, 
		p.channel_title, 
		p.channel_thumbnails,
		p.updated_at
	FROM playlist AS p
	JOIN post ON post.playlist_db_id = p.id
	GROUP BY p.id
	ORDER BY p.id DESC
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
