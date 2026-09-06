-- Create the post_credits table
CREATE TABLE post_credits (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(256) NOT NULL,
    role       VARCHAR(256) NOT NULL CHECK (role IN ('Director', 'Producer', 'Editor')),
    search_vector tsvector, -- search vector column
    post_id    INTEGER NOT NULL REFERENCES post(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (post_id, role, name)
);


-- Create GIN index on the post_credits search_vector column
CREATE INDEX idx_post_credits_search_vector ON post_credits USING GIN (search_vector);


-- Create GIN index on the post_credits name column for the pg_trgm
CREATE INDEX idx_post_credits_name_trgm ON post_credits USING GIN (name gin_trgm_ops);


-- Create trigger on the post_credits table to update the updated_at timestamp
CREATE TRIGGER post_credits_timestamp_update
    BEFORE UPDATE ON post_credits
    FOR EACH ROW EXECUTE FUNCTION update_timestamp();


-- Function to update the post_credits search_vector value
CREATE FUNCTION update_post_credits_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector =
        setweight(to_tsvector('english', coalesce(NEW.name, '')), 'A');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


-- Create a trigger on the post_review table to update the search_vector value
CREATE TRIGGER post_credits_tsvector_update
BEFORE INSERT OR UPDATE OF name ON post_credits
FOR EACH ROW EXECUTE FUNCTION update_post_credits_search_vector();


-- Add release_year column in the post table
ALTER TABLE post ADD COLUMN release_year SMALLINT;
