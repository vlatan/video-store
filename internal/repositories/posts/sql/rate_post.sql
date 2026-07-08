INSERT INTO post_rating (rating, user_id, post_id)
SELECT $1, $2, p.id
FROM post AS p
WHERE p.video_id = $3
ON CONFLICT (user_id, post_id)
DO UPDATE SET rating = EXCLUDED.rating;
