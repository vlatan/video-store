-- Remove the release_year column from post
ALTER TABLE post DROP COLUMN release_year;


-- Drop the tsvector trigger and its function
DROP TRIGGER post_credits_tsvector_update ON post_credits;
DROP FUNCTION update_post_credits_search_vector();


-- Drop the updated_at trigger
DROP TRIGGER post_credits_timestamp_update ON post_credits;


-- Drop indexes
DROP INDEX idx_post_credits_name_trgm;
DROP INDEX idx_post_credits_search_vector;


-- Drop the table
DROP TABLE post_credits;
