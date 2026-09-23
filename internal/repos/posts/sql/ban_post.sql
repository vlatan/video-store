WITH dp AS (
    DELETE FROM post
    WHERE video_id = $1
    RETURNING video_id, NULLIF(provider, '') as provider
)
INSERT INTO deleted_post (video_id, provider)
SELECT video_id, provider FROM dp;
