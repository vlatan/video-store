
WITH updated_post AS (
    UPDATE post
    SET
        original_title = $2,
        category_id = (SELECT id FROM category WHERE name = $3),
        summary = $4,
        release_year = $5
    WHERE video_id = $1
    RETURNING id
),
delete_credits AS (
    DELETE FROM post_credits AS pc
    USING updated_post AS up
    WHERE pc.post_id = up.id
)
INSERT INTO post_credits (post_id, name, role)
SELECT up.id, c.name, c.role 
FROM updated_post AS up
CROSS JOIN UNNEST($6::varchar(256)[], $7::varchar(256)[]) AS c(name, role)
