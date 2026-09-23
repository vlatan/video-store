WITH target_post AS (
    SELECT id AS post_id
    FROM post
    WHERE video_id = $1
),
agg_likes AS (
    SELECT
        tp.post_id,
        COUNT(*) AS likes_count
    FROM post_like AS pl
    JOIN target_post AS tp ON pl.post_id = tp.post_id
    GROUP BY tp.post_id
),
agg_rating AS (
    SELECT
        tp.post_id,
        ROUND(AVG(prat.rating), 2)::float8 AS avg_rating,
        COUNT(prat.rating) AS rating_count
    FROM post_rating AS prat
    JOIN target_post AS tp ON prat.post_id = tp.post_id
    GROUP BY tp.post_id
),
agg_directors AS (
    SELECT
        tp.post_id,
        ARRAY_AGG(pc.name) AS directors
    FROM post_credits AS pc
    JOIN target_post AS tp ON pc.post_id = tp.post_id
    WHERE pc.role = 'Director'
    GROUP BY tp.post_id
)
SELECT
    post.id,
    post.video_id,
    post.title,
    post.original_title,
    post.thumbnails,
    COALESCE(al.likes_count, 0) AS likes,
    arat.avg_rating,
    COALESCE(arat.rating_count, 0) AS rating_count,
    COALESCE(ad.directors, '{}') AS directors,
    post.description,
    post.summary,
    post.release_year,
    source.playlist_id,
    source.title,
    source.channel_title,
    cat.slug,
    cat.name,
    post.upload_date,
    post.duration
FROM post
JOIN target_post AS tp ON tp.post_id = post.id
LEFT JOIN agg_likes AS al ON al.post_id = post.id
LEFT JOIN agg_rating AS arat ON arat.post_id = post.id
LEFT JOIN agg_directors AS ad ON ad.post_id = post.id
LEFT JOIN category AS cat ON cat.id = post.category_id
LEFT JOIN playlist AS source ON source.id = post.playlist_db_id;
