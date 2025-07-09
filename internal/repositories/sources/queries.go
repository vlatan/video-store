package sources

const getSourcesQuery = `
	SELECT playlist_id, title, channel_title, channel_thumbnails FROM playlist
	WHERE id IN (SELECT DISTINCT playlist_db_id FROM post)
	ORDER BY id DESC
`
