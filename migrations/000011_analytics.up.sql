-- Rename column and make it not null.
-- WARNING: This needs to follow a code change that will use the new column name.
ALTER TABLE app_user RENAME COLUMN analytics_id TO public_id;
ALTER TABLE app_user ALTER COLUMN public_id SET NOT NULL;
