SELECT
    provider_user_id,
    provider, 
    name,
    email,
    picture,
    analytics_id,
    last_seen,
    created_at,
    COUNT(*) OVER() as total_results
FROM app_user
ORDER BY created_at
LIMIT $1 OFFSET $2;
