-- This will cascade and delete the review associated with this rating too.
DELETE FROM post_rating 
USING post AS p 
WHERE post_rating.post_id = p.id 
AND post_rating.user_id = $1 
AND p.video_id = $2
RETURNING p.id;
