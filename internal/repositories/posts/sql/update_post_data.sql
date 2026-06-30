UPDATE post
SET
    original_title = $2,
    category_id = (SELECT id FROM category WHERE name = $3),
    summary = $4
WHERE video_id = $1;
