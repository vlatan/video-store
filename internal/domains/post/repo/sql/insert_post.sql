WITH deleted_rows AS (
    DELETE FROM deleted_post
    WHERE video_id = $1
),
inserted_post AS (
    INSERT INTO post (
        video_id, 
        provider,
        playlist_id, 
        title,
        original_title,
        release_year,
        thumbnails, 
        description, 
        summary,
        tags, 
        duration, 
        upload_date, 
        user_id,
        category_id,
        playlist_db_id
    )
    VALUES (
        $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,
        (SELECT id FROM category WHERE name = $14),
        (SELECT id FROM playlist WHERE playlist_id = $3::varchar(50))
    )
    RETURNING id
),
inserted_credits AS (
    INSERT INTO post_credits (post_id, name, role)
    SELECT inserted_post.id, c.name, c.role
    FROM inserted_post
    CROSS JOIN UNNEST($15::varchar(256)[], $16::varchar(256)[]) AS c(name, role)
)
SELECT id FROM inserted_post;
