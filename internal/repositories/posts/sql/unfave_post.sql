DELETE FROM post_fave 
USING post AS p 
WHERE post_fave.post_id = p.id 
AND post_fave.user_id = $1 
AND p.video_id = $2;
