-- This migration is tied with a code change.
-- Namely using the 'summary' column instead of 'short_description'.

-- Rename the column
ALTER TABLE post 
RENAME COLUMN short_description TO summary;

-- Update the search vector function.
-- We use CREATE OR REPLACE to overwrite the existing logic.
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