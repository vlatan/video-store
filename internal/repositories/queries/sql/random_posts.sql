SELECT
    p.video_id,
    p.title,
    p.original_title,
    p.thumbnails,
    COUNT(pl.id) AS likes
FROM post AS p
LEFT JOIN post_like AS pl ON pl.post_id = p.id
WHERE p.title != $1 AND p.original_title != $1
GROUP BY p.id
ORDER BY RANDOM()
LIMIT $2;
