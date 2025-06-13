-- Used fo fuzzy search
CREATE EXTENSION IF NOT EXISTS pg_trgm;


-- Create search_vector column
ALTER TABLE post ADD COLUMN IF NOT EXISTS search_vector tsvector;


-- Create function to automatically update the search_vector column
CREATE OR REPLACE FUNCTION update_post_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector =
        setweight(to_tsvector('english', coalesce(NEW.title, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(NEW.short_description, '')), 'B') ||
        setweight(to_tsvector('english', coalesce(NEW.description, '')), 'C') ||
        setweight(to_tsvector('english', coalesce(NEW.tags, '')), 'D');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


-- Create a trigger to use the function above on post insert or update
CREATE OR REPLACE TRIGGER tsvector_update
BEFORE INSERT OR UPDATE ON post
FOR EACH ROW EXECUTE FUNCTION update_post_search_vector();


-- Create GIN index on the search_vector column
CREATE INDEX IF NOT EXISTS idx_post_search_vector ON post USING GIN (search_vector);


-- Create GIN index on the title column for the pg_trgm
CREATE INDEX IF NOT EXISTS idx_post_title_trgm ON post USING GIN (title gin_trgm_ops);


-- Populate the search_vector column for the existing data
UPDATE post SET search_vector =
    setweight(to_tsvector('english', coalesce(title, '')), 'A') ||
    setweight(to_tsvector('english', coalesce(short_description, '')), 'B') ||
    setweight(to_tsvector('english', coalesce(description, '')), 'C') ||
    setweight(to_tsvector('english', coalesce(tags, '')), 'D');