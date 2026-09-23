SELECT 
    prev.rating_id,
    prev.title AS headline,
    prev.review AS content,
    prat.rating AS rating,
    GREATEST(prat.updated_at, prev.updated_at) AS updated_at,
    prev.created_at,
    {{ .TotalCount }} AS total_results,
    au.id as user_id,
    au.provider_user_id,
    au.provider,
    au.name,
    au.email,
    au.picture,
    au.public_id
FROM post AS p
JOIN post_rating AS prat ON prat.post_id = p.id
JOIN post_review AS prev ON prev.rating_id = prat.id
JOIN app_user AS au ON au.id = prat.user_id
WHERE p.video_id = $1
{{ .AndCondition }} -- the AND condition if any
ORDER BY prev.created_at DESC, prev.rating_id DESC
LIMIT $2;
