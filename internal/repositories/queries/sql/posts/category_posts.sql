WITH posts_with_likes AS (
    SELECT 
        c.name AS category_title,
        post.id,
        post.video_id, 
        post.title,
        post.original_title,
        post.thumbnails,
        COUNT(pl.id) AS likes,
        {{ .TotalCount }} AS total_results,
        post.upload_date
    FROM post
    JOIN category AS c ON c.id = post.category_id 
    LEFT JOIN post_like AS pl ON pl.post_id = post.id
    WHERE c.slug = $1
    GROUP BY c.id, post.id
)
SELECT * FROM posts_with_likes
{{ .WhereCondition }} -- the WHERE condition if any
ORDER BY {{ .OrderByWhat }}
LIMIT $2;
