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
        p.title AS playlist_title,
        post.id,
        video_id, 
        post.title,
        original_title,
        post.thumbnails,
        COALESCE(l.likes, 0) AS likes,
        r.avg_rating,
        COALESCE(r.rating_count, 0) AS rating_count,
        {{ .TotalCount }} AS total_results,
        upload_date
    FROM post
    LEFT JOIN playlist AS p ON p.id = post.playlist_db_id 
    LEFT JOIN likes AS l ON l.post_id = post.id
    LEFT JOIN ratings AS r ON r.post_id = post.id
    WHERE
        CASE 
            WHEN $1 = 'other'
            THEN (p.playlist_id IS NULL OR p.playlist_id = '')
            ELSE p.playlist_id = $1
        END
)
SELECT * FROM posts
{{ .WhereCondition }} -- the WHERE condition if any
ORDER BY {{ .OrderByWhat }}
LIMIT $2;
