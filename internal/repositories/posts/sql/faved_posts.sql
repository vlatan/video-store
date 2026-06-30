WITH faved_posts AS (
    SELECT
        p.id,
        p.video_id,
        p.title,
        p.original_title,
        p.thumbnails,
        COUNT(pl.id) AS likes,
        {{ .TotalCount }} AS total_results,
        p.upload_date,
        pf.created_at AS when_faved
    FROM post AS p
    LEFT JOIN post_like AS pl ON pl.post_id = p.id
    LEFT JOIN post_fave AS pf ON pf.post_id = p.id
    WHERE pf.user_id = $1
    GROUP BY p.id, pf.id
)
SELECT * FROM faved_posts
{{ .WhereCondition }} -- the WHERE condition if any
ORDER BY when_faved DESC, likes DESC, upload_date DESC, id DESC
LIMIT $2;