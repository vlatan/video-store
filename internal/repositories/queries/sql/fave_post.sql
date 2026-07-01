INSERT INTO post_fave (user_id, post_id)
SELECT $1, p.id 
FROM post AS p 
WHERE p.video_id = $2;
