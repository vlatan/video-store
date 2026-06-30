    -- Posts (last modified = last updated_at)
	SELECT
		'post' as type,
		(id % $1) AS bucket_id,
		CONCAT('/video/', video_id, '/') AS location,
		updated_at AS last_modified
	FROM post

	UNION ALL

	-- Pages (last modified = last updated_at)
	SELECT
		'misc' AS type,
		0 AS bucket_id,
		CONCAT('/page/', slug, '/') AS location,
		updated_at AS last_modified
	FROM page

	UNION ALL

	-- Playlists (last modified = latest upload date post in playlist)
	SELECT
		'misc' AS type,
		0 AS bucket_id,
		CONCAT('/source/', p.playlist_id, '/') AS location, 
		MAX(post.upload_date) AS last_modified
	FROM playlist AS p
	INNER JOIN post ON post.playlist_db_id = p.id
	GROUP BY p.id

	UNION ALL

	-- Orphans (last modified = latest upload date post without playlist)
	SELECT
		'misc' AS type,
		0 AS bucket_id,
		'/source/other/' AS location,
		MAX(upload_date) AS last_modified
	FROM post
	WHERE playlist_id IS NULL OR playlist_id = ''
	HAVING COUNT(*) > 0

	UNION ALL

	-- Categories (last modified = latest upload date post in category)
	SELECT
		'misc' AS type,
		0 AS bucket_id,
		CONCAT('/category/', c.slug, '/') AS location,
		MAX(post.upload_date) AS last_modified
	FROM category AS c
	INNER JOIN post ON post.category_id = c.id
	GROUP BY c.id

	UNION ALL

	-- Homepage (last modified = latest upload date post in DB)
	SELECT
		'misc' AS type,
		0 AS bucket_id,
		'/' AS location,
		MAX(upload_date) AS last_modified
	FROM post

	UNION ALL

	-- Playlists page (last modified = newest playlist in DB)
	SELECT 
		'misc' AS type,
		0 AS bucket_id,
		'/sources/' AS location, 
		MAX(created_at) AS last_modified
	FROM playlist

	ORDER BY type, location;