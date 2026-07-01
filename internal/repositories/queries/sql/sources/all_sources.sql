SELECT
    playlist_id,
    channel_id,
    title, 
    channel_title, 
    channel_thumbnails,
    updated_at
FROM playlist
ORDER BY id DESC;
