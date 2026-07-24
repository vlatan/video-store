SELECT 
    au.provider_user_id,
    au.provider,
    au.name,
    au.email,
    au.picture,
    au.analytics_id,
    prev.title AS headline,
    prev.review AS content,
    prat.rating AS rating,
    GREATEST(prat.updated_at, prev.updated_at) AS updated_at
FROM post AS p
JOIN post_rating AS prat ON prat.post_id = p.id
JOIN post_review AS prev ON prev.rating_id = prat.id
JOIN app_user AS au ON au.id = prat.user_id
WHERE p.video_id = $1
ORDER BY prev.created_at DESC;
