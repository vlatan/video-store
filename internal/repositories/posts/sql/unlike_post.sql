DELETE FROM post_like 
USING post AS p 
WHERE post_like.post_id = p.id 
AND post_like.user_id = $1 
AND p.video_id = $2;
