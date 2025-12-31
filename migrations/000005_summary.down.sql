-- This migration is tied with a code change.
-- Namely using the 'short_description' column instead of 'summary'.

-- Rename the column back
ALTER TABLE post 
RENAME COLUMN summary TO short_description;

-- Revert the search vector function.
-- We point it back to the old column name.
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