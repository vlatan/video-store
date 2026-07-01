WITH posts_with_likes AS (
    SELECT 
        p.title AS playlist_title,
        post.id,
        post.video_id, 
        post.title,
        post.original_title,
        post.thumbnails,
        COUNT(pl.id) AS likes,
        {{ .TotalCount }} AS total_results,
        post.upload_date
    FROM post
    LEFT JOIN playlist AS p ON p.id = post.playlist_db_id 
    LEFT JOIN post_like AS pl ON pl.post_id = post.id
    WHERE
        CASE 
            WHEN $1 = 'other'
            THEN (p.playlist_id IS NULL OR p.playlist_id = '')
            ELSE p.playlist_id = $1
        END
    GROUP BY p.id, post.id
)
SELECT * FROM posts_with_likes
{{ .WhereCondition }} -- the WHERE condition if any
ORDER BY {{ .OrderByWhat }}
LIMIT $2;
