WITH deleted_rows AS (
    DELETE FROM deleted_post
    WHERE video_id = $1
)
INSERT INTO post (
    video_id, 
    provider,
    playlist_id, 
    title,
    original_title,
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
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12,
    (SELECT id FROM category WHERE name = $13),
    (SELECT id FROM playlist WHERE playlist_id = $3::varchar(50))
);
