-- Insert or update a rating + review together
-- (posting a review always requires a rating)
WITH target_post AS (
    SELECT id FROM post WHERE video_id = $3
),
rating_upsert AS (
    INSERT INTO post_rating (rating, user_id, post_id)
    SELECT $1, $2, tp.id
    FROM target_post AS tp
    ON CONFLICT (user_id, post_id)
    DO UPDATE SET rating = EXCLUDED.rating
    RETURNING id
)
INSERT INTO post_review (rating_id, title, review)
SELECT ru.id, $4, $5
FROM rating_upsert AS ru
ON CONFLICT (rating_id)
DO UPDATE SET
    title = EXCLUDED.title, 
    review = EXCLUDED.review
RETURNING rating_id;

