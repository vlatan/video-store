-- Add new columns
ALTER TABLE app_user 
ADD COLUMN provider_user_id VARCHAR(256),
ADD COLUMN provider VARCHAR(50);

-- Migrate existing Google users
UPDATE app_user 
SET provider_user_id = google_id, provider = 'google' 
WHERE google_id IS NOT NULL;

-- Migrate existing Facebook users
UPDATE app_user 
SET provider_user_id = facebook_id, provider = 'facebook' 
WHERE facebook_id IS NOT NULL;

-- Add unique constraint on the combination of provider_user_id and provider
-- This ensures a user can't have the same provider_user_id for the same provider
ALTER TABLE app_user
ALTER COLUMN provider_user_id SET NOT NULL,
ALTER COLUMN provider SET NOT NULL,
ADD CONSTRAINT app_user_provider_user_id_provider_key UNIQUE (provider_user_id, provider);