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
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULLIF($9, 0));
