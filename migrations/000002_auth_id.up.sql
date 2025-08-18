-- Add new columns
ALTER TABLE app_user 
ADD COLUMN auth_id VARCHAR(256),
ADD COLUMN provider VARCHAR(50);

-- Migrate existing Google users
UPDATE app_user 
SET auth_id = google_id, provider = 'google' 
WHERE google_id IS NOT NULL;

-- Migrate existing Facebook users
UPDATE app_user 
SET auth_id = facebook_id, provider = 'facebook' 
WHERE facebook_id IS NOT NULL;

-- Add unique constraint on the combination of auth_id and provider
-- This ensures a user can't have the same auth_id for the same provider
ALTER TABLE app_user
ALTER COLUMN auth_id SET NOT NULL,
ADD CONSTRAINT app_user_auth_id_provider_key UNIQUE (auth_id, provider);