WITH likes AS (
    SELECT post_id, COUNT(*) AS likes
    FROM post_like
    GROUP BY post_id
),
ratings AS (
    SELECT
        post_id,
        ROUND(AVG(rating), 2)::float8 AS avg_rating,
        COUNT(rating) AS rating_count
    FROM post_rating
    GROUP BY post_id
),
posts AS (
    SELECT
        post.id,
        video_id,
        title,
        original_title,
        thumbnails,
        COALESCE(l.likes, 0) AS likes,
        r.avg_rating,
        COALESCE(r.rating_count, 0) AS rating_count,
        {{ .TotalCount }} AS total_results,
        upload_date,
        pf.created_at AS when_faved
    FROM post
    JOIN post_fave AS pf ON pf.post_id = post.id
    LEFT JOIN likes AS l ON l.post_id = post.id
    LEFT JOIN ratings AS r ON r.post_id = post.id
    WHERE pf.user_id = $1
)
SELECT * FROM posts
{{ .WhereCondition }} -- the WHERE condition if any
ORDER BY when_faved DESC, upload_date DESC, id DESC
LIMIT $2;
