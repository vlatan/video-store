CREATE TABLE post_review (
    id SERIAL PRIMARY KEY,
    title VARCHAR(256),
    review TEXT,
    search_vector tsvector, -- search vector column
    user_id INTEGER REFERENCES app_user(id) ON DELETE CASCADE,
    post_id INTEGER REFERENCES post(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, post_id) -- Prevent duplicates
);


-- Create FK index on the post_review table for fast lookup for joins with post
CREATE INDEX idx_post_review_post_id ON post_review(post_id);


-- Create GIN index on the post_review search_vector column
CREATE INDEX idx_post_review_search_vector ON post_review USING GIN (search_vector);


-- Create GIN index on the post_review title column for the pg_trgm
CREATE INDEX idx_post_review_title_trgm ON post_review USING GIN (title gin_trgm_ops);


-- Create trigger on the post_review table to update the updated_at timestamp
CREATE TRIGGER post_review_timestamp_update
    BEFORE UPDATE ON post_review
    FOR EACH ROW EXECUTE FUNCTION update_timestamp();


-- Function to update the post_review search_vector value
CREATE FUNCTION update_post_review_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector =
        setweight(to_tsvector('english', coalesce(NEW.title, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(NEW.review, '')), 'B');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


-- Create a trigger on the post_review table to update the search_vector value
CREATE TRIGGER post_review_tsvector_update
BEFORE INSERT OR UPDATE OF title, review ON post_review
FOR EACH ROW EXECUTE FUNCTION update_post_review_search_vector();


CREATE TABLE post_rating (
    id SERIAL PRIMARY KEY,
    rating SMALLINT CHECK (rating >= 1 AND rating <= 10),
    user_id INTEGER REFERENCES app_user(id) ON DELETE CASCADE,
    post_id INTEGER REFERENCES post(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, post_id) -- Prevent duplicates
);


-- Create FK index on the post_rating table for fast lookup for joins with post
CREATE INDEX idx_post_rating_post_id ON post_rating(post_id);


-- Create trigger on the post_rating table to update the updated_at timestamp
CREATE TRIGGER post_rating_timestamp_update
    BEFORE UPDATE ON post_rating
    FOR EACH ROW EXECUTE FUNCTION update_timestamp();



--------- CHANGES UNRELATED TO REVIEWS OR RATINGS TO FURTHER OPTIMISE THE DATABASE ---------

-- Create FK index on the post_like table for fast lookup for joins with post
CREATE INDEX idx_post_like_post_id ON post_like(post_id);


-- Create FK index on the post_fave table for fast lookup for joins with post
CREATE INDEX idx_post_fave_post_id ON post_fave(post_id);


-- Update the trigger on the post table to update the search_vector value
-- only on just the relevant columns for the search.
CREATE OR REPLACE TRIGGER post_tsvector_update
BEFORE INSERT OR UPDATE OF title, original_title, summary, description, tags ON post
FOR EACH ROW EXECUTE FUNCTION update_post_search_vector();
