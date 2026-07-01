INSERT INTO app_user (
    provider_user_id, 
    provider, 
    analytics_id, 
    name, 
    email, 
    picture, 
    last_seen
)
VALUES ( $1, $2, $3, $4, $5, $6, NOW() ) 
ON CONFLICT (provider, provider_user_id) 
DO UPDATE SET
    analytics_id = EXCLUDED.analytics_id,
    name = EXCLUDED.name,
    email = EXCLUDED.email,
    picture = EXCLUDED.picture,
    last_seen = NOW()
RETURNING id;
