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
		category_id
	)
	VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NULLIF($11, 0),
		(SELECT id FROM category WHERE name = $12)
	)
`

const getPostsQuery = `
	SELECT video_id, title, thumbnails, (
		SELECT COUNT(*) FROM post_like
		WHERE post_like.post_id = post.id
	) AS likes FROM post
	ORDER BY %s
	LIMIT $1 OFFSET $2
`

const getSinglePostQuery = `
	SELECT 
		post.id,
		video_id,
		title, 
		thumbnails, (
			SELECT COUNT(*) FROM post_like
			WHERE post_like.post_id = post.id
		) AS likes, 
		description,
		short_description,
		slug AS category_slug,
		name AS category_name,
		upload_date,
		duration
	FROM post 
	LEFT JOIN category ON post.category_id = category.id
	WHERE video_id = $1 
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
	LIMIT $2 OFFSET $3;
`

const getSourcePostsQuery = `
	SELECT 
		p.title AS playlist_title, 
		post.video_id, 
		post.title, 
		post.thumbnails,
		COUNT(pl.id) AS likes
	FROM post
	JOIN playlist AS p ON post.playlist_db_id = p.id
	LEFT JOIN post_like AS pl ON pl.post_id = post.id
	WHERE p.playlist_id = $1
	GROUP BY p.id, post.id
	ORDER BY %s
	LIMIT $2 OFFSET $3;
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
	SELECT video_id, provider FROM dp;
`
