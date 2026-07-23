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
agg_reviews AS (
    SELECT 
        tp.post_id,
        JSON_AGG(JSON_BUILD_OBJECT(
            'username', au.name,
            'headline', prev.title,
            'content', prev.review,
            'rating', prat.rating,
            -- Force PostgreSQL to append "Z" so Go's JSON parser recognizes it as UTC
            'updated_at', to_char(GREATEST(prat.updated_at, prev.updated_at), 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
        )) AS reviews_json
    FROM post_review AS prev
    JOIN post_rating AS prat ON prat.id = prev.rating_id
    JOIN target_post AS tp ON prat.post_id = tp.post_id
    JOIN app_user AS au ON au.id = prat.user_id
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
    COALESCE(arev.reviews_json, '[]') AS reviews,
    post.description,
    post.summary,
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
LEFT JOIN agg_reviews AS arev ON arev.post_id = post.id
LEFT JOIN category AS cat ON cat.id = post.category_id
LEFT JOIN playlist AS source ON source.id = post.playlist_db_id;
