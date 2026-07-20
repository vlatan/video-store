-- Insert or update user post review
INSERT INTO post_review (title, review, user_id, post_id)
SELECT $1, $2, $3, p.id
FROM post AS p
WHERE p.video_id = $3
ON CONFLICT (user_id, post_id)
DO UPDATE SET
    title = EXCLUDED.title,
    review = EXCLUDED.review
RETURNING post_id;
