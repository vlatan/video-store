package posts

const postExistsQuery = `
	SELECT 1 FROM post
	WHERE video_id = $1
`

const insertPostQuery = `
	WITH deleted_rows AS (
		DELETE FROM deleted_post
		WHERE video_id = $1
	)
	INSERT INTO post (
		video_id, 
		provider,
		playlist_id, 
		title, 
		thumbnails, 
		description, 
		short_description,
		tags, 
		duration, 
		upload_date, 
		user_id,
		category_id,
		playlist_db_id
	)
	VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NULLIF($11, 0),
		(SELECT id FROM category WHERE name = $12),
		(SELECT id FROM playlist WHERE playlist_id = $3::varchar(50))
	)
`

const getSinglePostQuery = `
	SELECT 
		post.id,
		post.video_id,
		post.title, 
		post.thumbnails,
		COUNT(pl.id) AS likes,
		post.description,
		post.short_description,
		category.slug,
		category.name,
		post.upload_date,
		post.duration
	FROM post 
	LEFT JOIN post_like AS pl ON pl.post_id = post.id
	LEFT JOIN category ON category.id = post.category_id
	WHERE video_id = $1
	GROUP BY post.id, category.id
`

const getPostsQuery = `
	SELECT
		video_id, 
		title, 
		thumbnails,
		COUNT(pl.id) AS likes
	FROM post
	LEFT JOIN post_like AS pl ON pl.post_id = post.id
	GROUP BY post.id
	ORDER BY %s
	LIMIT $1 OFFSET $2
`

const getAllPostsQuery = `
	SELECT
		video_id,
		playlist_id,
		title, 
		short_description
	FROM post
`

const getCategoryPostsQuery = `
	SELECT 
		c.name AS category_title, 
		post.video_id, 
		post.title, 
		post.thumbnails,
		COUNT(pl.id) AS likes
	FROM post
	JOIN category AS c ON c.id = post.category_id 
	LEFT JOIN post_like AS pl ON pl.post_id = post.id
	WHERE c.slug = $1
	GROUP BY c.id, post.id
	ORDER BY %s
	LIMIT $2 OFFSET $3
`

const getSourcePostsQuery = `
	SELECT 
		p.title AS playlist_title, 
		post.video_id, 
		post.title, 
		post.thumbnails,
		COUNT(pl.id) AS likes
	FROM post
	LEFT JOIN playlist AS p ON post.playlist_db_id = p.id
	LEFT JOIN post_like AS pl ON pl.post_id = post.id
	WHERE
		CASE 
    		WHEN $1 = 'other'
			THEN (p.playlist_id IS NULL OR p.playlist_id = '')
    		ELSE p.playlist_id = $1
  		END
	GROUP BY p.id, post.id
	ORDER BY %s
	LIMIT $2 OFFSET $3
`

const getUserFavedPostsQuery = `
	SELECT
		p.video_id,
		p.title,
		p.thumbnails,
		COUNT(pl.id) AS likes,
		CASE 
			WHEN $3 = 0 THEN COUNT(*) OVER()
			ELSE 0
		END AS total_results
	FROM post AS p
	LEFT JOIN post_like AS pl ON pl.post_id = p.id
	LEFT JOIN post_fave AS pf ON pf.post_id = p.id
	WHERE pf.user_id = $1
	GROUP BY p.id, pf.id
	ORDER BY pf.created_at, p.upload_date
	LIMIT $2 OFFSET $3
`

const searchPostsQuery = `
	WITH search_terms AS (
		SELECT
			lexeme AS and_query,
			to_tsquery('english', replace(lexeme::text, ' & ', ' | ')) AS or_query,
			replace(lexeme::text, ' & ', ' ') AS raw_query
		FROM plainto_tsquery('english', $1) AS lexeme
	)
	SELECT
		p.video_id,
		p.title,
		p.thumbnails,
		COUNT(pl.id) AS likes,
		CASE 
			WHEN $3 = 0 THEN COUNT(*) OVER()
			ELSE 0
		END AS total_results
	FROM post AS p
	CROSS JOIN search_terms AS st
	LEFT JOIN post_like AS pl ON pl.post_id = p.id
	WHERE p.search_vector @@ st.and_query OR p.search_vector @@ st.or_query
	GROUP BY p.id, st.and_query, st.or_query, st.raw_query
	ORDER BY 
		(ts_rank(p.search_vector, st.and_query, 32) * 2) + 
		ts_rank(p.search_vector, st.or_query, 32) +
		(similarity(p.title, st.raw_query) * 0.5) DESC,
		likes DESC,
		p.upload_date DESC
	LIMIT $2 OFFSET $3
`

const userActionsQuery = `
	SELECT 
		EXISTS (
			SELECT 1 FROM post_like
			WHERE user_id = $1 AND post_id = $2
		) AS liked,
		EXISTS (
			SELECT 1 FROM post_fave
			WHERE user_id = $1 AND post_id = $2
		) AS faved
`

const likeQuery = `
	INSERT INTO post_like (user_id, post_id)
	SELECT $1, p.id 
	FROM post AS p 
	WHERE p.video_id = $2
`

const unlikeQuery = `
	DELETE FROM post_like 
	USING post AS p 
	WHERE post_like.post_id = p.id 
	AND post_like.user_id = $1 
	AND p.video_id = $2
`

const faveQuery = `
	INSERT INTO post_fave (user_id, post_id)
	SELECT $1, p.id 
	FROM post AS p 
	WHERE p.video_id = $2
`

const unfaveQuery = `
	DELETE FROM post_fave 
	USING post AS p 
	WHERE post_fave.post_id = p.id 
	AND post_fave.user_id = $1 
	AND p.video_id = $2
`

const updateTitleQuery = `
	UPDATE post
	SET title = $2, updated_at = NOW()
	WHERE video_id = $1
`

const updateDescQuery = `
	UPDATE post
	SET short_description = $2, updated_at = NOW()
	WHERE video_id = $1
`

const deletePostQuery = `
	WITH dp AS (
		DELETE FROM post
		WHERE video_id = $1
		RETURNING video_id, NULLIF(provider, '') as provider
	)
	INSERT INTO deleted_post (video_id, provider)
	SELECT video_id, provider FROM dp
`

const isPostDeletedQuery = `
	SELECT 1 FROM deleted_post
	WHERE video_id = $1
`

const sitemapDataQuery = `
	-- Posts (last modified = last updated_at)
	SELECT
		'post' as type,
		CONCAT('/video/', video_id, '/') AS url,
		updated_at
	FROM post

	UNION ALL

	-- Pages (last modified = last updated_at)
	SELECT
		'page' AS type,
		CONCAT('/page/', slug, '/') AS url,
		updated_at
	FROM page

	UNION ALL

	-- Playlists (last modified = latest upload date post in playlist)
	SELECT
		'source' AS type, 
		CONCAT('/source/', p.playlist_id, '/') AS url, 
		MAX(post.upload_date) AS updated_at
	FROM playlist AS p
	LEFT JOIN post ON post.playlist_db_id = p.id
	GROUP BY p.id

	UNION ALL

	-- Orphans (last modified = latest upload date post without playlist)
	SELECT
		'source' AS type,
		'/source/other/' AS url,
		MAX(post.upload_date) AS updated_at
	FROM post
	WHERE playlist_id IS NULL OR playlist_id = ''

	UNION ALL

	-- Categories (last modified = latest upload date post in category)
	SELECT
		'category' AS type,
		CONCAT('/category/', c.slug, '/') AS url,
		MAX(post.upload_date) AS updated_at
	FROM category AS c
	LEFT JOIN post ON post.category_id = c.id
	GROUP BY c.id

	UNION ALL

	-- Homepage (last modified = latest upload date post in DB)
	SELECT
		'misc' AS type,
		'/' AS url,
		MAX(upload_date) AS updated_at
	FROM post

	UNION ALL

	-- Playlists page (last modified = newest playlist in DB)
	SELECT 
		'misc' AS type,
		'/sources/' AS url, 
		MAX(created_at) AS updated_at
	FROM playlist
`
