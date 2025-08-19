-- Remove the unique constraint
ALTER TABLE app_user DROP CONSTRAINT IF EXISTS app_user_provider_user_id_provider_key;

-- Drop the new columns
ALTER TABLE app_user DROP COLUMN IF EXISTS provider_user_id;
ALTER TABLE app_user DROP COLUMN IF EXISTS provider;