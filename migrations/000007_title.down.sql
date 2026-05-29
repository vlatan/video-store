-- Revert the search vector function.
-- We point it back to the old columns.
CREATE OR REPLACE FUNCTION update_post_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector =
        setweight(to_tsvector('english', coalesce(NEW.title, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(NEW.summary, '')), 'B') ||
        setweight(to_tsvector('english', coalesce(NEW.description, '')), 'C') ||
        setweight(to_tsvector('english', coalesce(NEW.tags, '')), 'D');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


-- Drop the index we created
DROP INDEX idx_post_original_title_trgm;


-- Drop the new column
ALTER TABLE post DROP COLUMN original_title;
