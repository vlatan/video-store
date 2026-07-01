UPDATE playlist
SET
    channel_id = $2,
    title = $3,
    channel_title = $4,
    thumbnails = $5,
    channel_thumbnails = $6,
    description = $7,
    channel_description = $8
WHERE playlist_id = $1;
