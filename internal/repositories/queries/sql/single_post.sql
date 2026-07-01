SELECT 
    post.id,
    post.video_id,
    post.title,
    post.original_title,
    post.thumbnails,
    COUNT(pl.id) AS likes,
    post.description,
    post.summary,
    playlist.playlist_id,
    playlist.title,
    playlist.channel_title,
    category.slug,
    category.name,
    post.upload_date,
    post.duration
FROM post 
LEFT JOIN post_like AS pl ON pl.post_id = post.id
LEFT JOIN category ON category.id = post.category_id
LEFT JOIN playlist ON playlist.id = post.playlist_db_id
WHERE video_id = $1
GROUP BY post.id, category.id, playlist.id;
