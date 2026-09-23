SELECT
    post.id,
    video_id,
    playlist_id,
    title,
    original_title,
    summary,
    duration,
    upload_date,
    cat.name AS category_name
FROM post
LEFT JOIN category AS cat ON cat.id = post.category_id
ORDER BY upload_date DESC, post.id DESC;
