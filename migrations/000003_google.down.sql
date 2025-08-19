-- Bring back the odl columns
ALTER TABLE app_user 
ADD COLUMN google_id VARCHAR(256) UNIQUE,
ADD COLUMN facebook_id VARCHAR(256) UNIQUE;

-- Bring back the existing Google users
UPDATE app_user 
SET google_id = provider_user_id
WHERE provider = 'google';

-- Bring back the existing Facebook users
UPDATE app_user 
SET facebook_id = provider_user_id
WHERE provider = 'facebook';