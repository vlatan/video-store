SELECT
    id,
    provider_user_id,
    provider, 
    name,
    email,
    picture,
    public_id,
    last_seen,
    created_at
FROM app_user
WHERE id = $1;
