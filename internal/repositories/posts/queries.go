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
	SET title = $2
	WHERE video_id = $1
`

const updateDescQuery = `
	UPDATE post
	SET short_description = $2
	WHERE video_id = $1
`

const updatePlaylistQuery = `
	UPDATE post
	SET 
		playlist_id = $2,
		playlist_db_id = (SELECT id FROM playlist WHERE playlist_id = $2)
	WHERE video_id = $1
`

const updateGeneretedDataQuery = `
	UPDATE post
	SET
		category_id = (SELECT id FROM category WHERE name = $2),
		short_description = $3
	WHERE video_id = $1
`

const banPostQuery = `
	WITH dp AS (
		DELETE FROM post
		WHERE video_id = $1
		RETURNING video_id, NULLIF(provider, '') as provider
	)
	INSERT INTO deleted_post (video_id, provider)
	SELECT video_id, provider FROM dp
`

const isPostBanneddQuery = `
	SELECT 1 FROM deleted_post
	WHERE video_id = $1
`

const deletePostQuery = `
	DELETE FROM post
	WHERE video_id = $1
`
