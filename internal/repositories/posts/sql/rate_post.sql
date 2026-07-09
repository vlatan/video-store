-- Insert post rating and return the new average rating and count for that post
WITH ins AS (
    INSERT INTO post_rating (rating, user_id, post_id)
    SELECT $1, $2, p.id
    FROM post AS p
    WHERE p.video_id = $3
    ON CONFLICT (user_id, post_id)
    DO UPDATE SET rating = EXCLUDED.rating
    RETURNING post_id
)
SELECT
    ROUND(AVG(pr.rating), 2)::float8 AS avg_rating
    COUNT(pr.rating) AS rating_count
FROM ins AS i
JOIN post_rating AS pr ON pr.post_id = i.post_id
GROUP BY i.post_id;
