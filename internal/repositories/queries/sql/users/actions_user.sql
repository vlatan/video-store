SELECT 
    EXISTS (
        SELECT 1 FROM post_like
        WHERE user_id = $1 AND post_id = $2
    ) AS liked,
    EXISTS (
        SELECT 1 FROM post_fave
        WHERE user_id = $1 AND post_id = $2
    ) AS faved;
