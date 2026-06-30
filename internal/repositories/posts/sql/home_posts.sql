WITH posts_with_likes AS (
    SELECT
        post.id,
        video_id, 
        title,
        original_title, 
        thumbnails,
        COUNT(pl.id) AS likes,
        upload_date
    FROM post
    LEFT JOIN post_like AS pl ON pl.post_id = post.id
    GROUP BY post.id
)
SELECT * FROM posts_with_likes
{{ .WhereCondition }} -- the WHERE condition if any
ORDER BY {{ .OrderByWhat }}
LIMIT $1;