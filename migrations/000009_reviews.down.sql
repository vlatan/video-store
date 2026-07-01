-- Revert the trigger on the post table to its previous state (triggering on all updates)
CREATE OR REPLACE TRIGGER post_tsvector_update
BEFORE INSERT OR UPDATE ON post
FOR EACH ROW EXECUTE FUNCTION update_post_search_vector();


-- Drop indexes on existing tables
DROP INDEX IF EXISTS idx_post_fave_post_id;
DROP INDEX IF EXISTS idx_post_like_post_id;


-- Drop post_rating table (automatically drops its triggers and indexes)
DROP TABLE IF EXISTS post_rating;


-- Drop post_review table (automatically drops its triggers and indexes)
DROP TABLE IF EXISTS post_review;


-- Drop the search vector function created for post_review
DROP FUNCTION update_post_review_search_vector();