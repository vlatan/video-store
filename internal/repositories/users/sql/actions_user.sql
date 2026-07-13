SELECT
    p.user_id,
    p.post_id,
    pl.user_id IS NOT NULL AS liked,
    pf.user_id IS NOT NULL AS faved,
    pf.created_at AS when_faved,
    COALESCE(pr.rating, 0) AS rating
FROM (SELECT $1::integer AS user_id, $2::integer AS post_id) AS p
LEFT JOIN post_like AS pl ON pl.user_id = p.user_id AND pl.post_id = p.post_id
LEFT JOIN post_fave AS pf ON pf.user_id = p.user_id AND pf.post_id = p.post_id
LEFT JOIN post_rating AS pr ON pr.user_id = p.user_id AND pr.post_id = p.post_id;
