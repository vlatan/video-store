package posts

const postExistsQuery = `
	SELECT 1 FROM post
	WHERE video_id = $1
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
	SELECT video_id, title, thumbnails, (
		SELECT COUNT(*) FROM post_like
		WHERE post_like.post_id = post.id
	) AS likes FROM post 
	WHERE category_id = (SELECT id FROM category WHERE slug = $1) 
	ORDER BY %s
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
		(
			SELECT COUNT(*)
			FROM post_like
			WHERE post_like.post_id = p.id
		) AS likes,
		CASE 
			WHEN $3 = 0 THEN COUNT(*) OVER()
			ELSE 0
		END AS total_results
	FROM post AS p, search_terms AS st
	WHERE p.search_vector @@ st.and_query OR p.search_vector @@ st.or_query
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
		RETURNING video_id
	)
	INSERT INTO deleted_post (video_id)
	SELECT video_id FROM dp;
`
