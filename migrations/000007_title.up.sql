-- Add new column to post table
ALTER TABLE post ADD COLUMN original_title VARCHAR(256);


-- Create GIN index on the title column for the pg_trgm
CREATE INDEX idx_post_original_title_trgm ON post USING GIN (original_title gin_trgm_ops);


-- Update the search vector function.
-- We use CREATE OR REPLACE to overwrite the existing logic.
CREATE OR REPLACE FUNCTION update_post_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector =
        setweight(to_tsvector('english', coalesce(NEW.title, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(NEW.original_title, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(NEW.summary, '')), 'B') ||
        setweight(to_tsvector('english', coalesce(NEW.description, '')), 'C') ||
        setweight(to_tsvector('english', coalesce(NEW.tags, '')), 'D');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;