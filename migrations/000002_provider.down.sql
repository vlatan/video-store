-- There's a app code change already DONE.
-- If you run this expect the app to break.
-- The app expects to find the `provider` and `provider_user_id` columns.

-- Remove the unique constraint
ALTER TABLE app_user 
DROP CONSTRAINT app_user_provider_user_id_provider_key;

-- Drop the new columns
ALTER TABLE app_user
DROP COLUMN provider_user_id,
DROP COLUMN provider;