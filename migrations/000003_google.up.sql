-- Drop the old columns
ALTER TABLE app_user
DROP COLUMN google_id,
DROP COLUMN facebook_id;