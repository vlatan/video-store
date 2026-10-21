UPDATE post
SET 
    playlist_id = $2,
    playlist_db_id = (
        SELECT id
        FROM playlist
        WHERE playlist_id = $2::VARCHAR(50)
    )
WHERE video_id = $1;
