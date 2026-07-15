SELECT
    post.id,
    post.video_id,
    post.title,
    post.original_title,
    post.thumbnails,
    COALESCE(l.likes, 0) AS likes,
    r.avg_rating,
    COALESCE(r.rating_count, 0) AS rating_count,
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
LEFT JOIN LATERAL (
    SELECT COUNT(*) AS likes
    FROM post_like
    WHERE post_like.post_id = post.id
) AS l ON true
LEFT JOIN LATERAL (
    SELECT
        ROUND(AVG(rating), 2)::float8 AS avg_rating,
        COUNT(rating) AS rating_count
    FROM post_rating
    WHERE post_rating.post_id = post.id
) AS r ON true
LEFT JOIN category ON category.id = post.category_id
LEFT JOIN playlist ON playlist.id = post.playlist_db_id
WHERE post.video_id = $1;
