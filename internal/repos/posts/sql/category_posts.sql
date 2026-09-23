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
        c.name AS category_title,
        post.id,
        video_id, 
        title,
        original_title,
        thumbnails,
        COALESCE(l.likes, 0) AS likes,
        r.avg_rating,
        COALESCE(r.rating_count, 0) AS rating_count,
        {{ .TotalCount }} AS total_results,
        upload_date
    FROM post
    JOIN category AS c ON c.id = post.category_id 
    LEFT JOIN likes AS l ON l.post_id = post.id
    LEFT JOIN ratings AS r ON r.post_id = post.id
    WHERE c.slug = $1
)
SELECT * FROM posts
{{ .WhereCondition }} -- the WHERE condition if any
ORDER BY {{ .OrderByWhat }}
LIMIT $2;
