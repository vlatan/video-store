-- Drop non null constraint and rename back the column
-- WARNING: This needs to follow a code change that will use the old column name
ALTER TABLE app_user ALTER COLUMN public_id DROP NOT NULL;
ALTER TABLE app_user RENAME COLUMN public_id TO analytics_id;
