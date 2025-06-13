-- Drop the index on the search_vector column
DROP INDEX IF EXISTS idx_post_search_vector;


-- Drop the index on title column
DROP INDEX IF EXISTS idx_title_trgm;


-- Drop the function trigger
DROP TRIGGER IF EXISTS tsvector_update ON post;


-- Remove the update_post_search_vector function
DROP FUNCTION IF EXISTS update_post_search_vector();


-- Drop the column search_vector
ALTER TABLE post DROP COLUMN IF EXISTS search_vector;


-- Remove the fuzzy search extension
DROP EXTENSION IF EXISTS pg_trgm;