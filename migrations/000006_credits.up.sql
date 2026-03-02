-- Add new columns to post table
ALTER TABLE post
    ADD COLUMN past_context TEXT,
    ADD COLUMN present_context TEXT,
    ADD COLUMN release_year INT,
    ADD COLUMN country_of_origin TEXT,
    ADD COLUMN language TEXT;

-- Update the post search vector function
CREATE OR REPLACE FUNCTION update_post_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector =
        setweight(to_tsvector('english', coalesce(NEW.title, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(NEW.summary, '')), 'B') ||
        setweight(to_tsvector('english', coalesce(NEW.past_context, '')), 'C') ||
        setweight(to_tsvector('english', coalesce(NEW.present_context, '')), 'C') ||
        setweight(to_tsvector('english', coalesce(NEW.description, '')), 'C') ||
        setweight(to_tsvector('english', coalesce(NEW.tags, '')), 'D');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Person table
CREATE TABLE person (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    bio TEXT,
    search_vector TSVECTOR,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE OR REPLACE FUNCTION update_person_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector =
        setweight(to_tsvector('english', coalesce(NEW.name, '')), 'C') ||
        setweight(to_tsvector('english', coalesce(NEW.bio, '')), 'D');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER person_search_vector_update
    BEFORE INSERT OR UPDATE ON person
    FOR EACH ROW
    EXECUTE FUNCTION update_person_search_vector();

CREATE TRIGGER person_timestamp_update
    BEFORE UPDATE ON person
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamp();

-- Production company table
CREATE TABLE production_company (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    search_vector TSVECTOR,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE OR REPLACE FUNCTION update_production_company_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector =
        setweight(to_tsvector('english', coalesce(NEW.name, '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER production_company_search_vector_update
    BEFORE INSERT OR UPDATE ON production_company
    FOR EACH ROW
    EXECUTE FUNCTION update_production_company_search_vector();

CREATE TRIGGER production_company_timestamp_update
    BEFORE UPDATE ON production_company
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamp();

-- Post person junction table
CREATE TABLE post_person (
    id SERIAL PRIMARY KEY,
    post_id INT NOT NULL REFERENCES post(id) ON DELETE CASCADE,
    person_id INT NOT NULL REFERENCES person(id) ON DELETE CASCADE,
    role TEXT NOT NULL, -- 'director', 'writer', 'narrator', 'appearance'
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(post_id, person_id, role)
);

CREATE TRIGGER post_person_timestamp_update
    BEFORE UPDATE ON post_person
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamp();

-- Post production company junction table
CREATE TABLE post_production_company (
    id SERIAL PRIMARY KEY,
    post_id INT NOT NULL REFERENCES post(id) ON DELETE CASCADE,
    production_company_id INT NOT NULL REFERENCES production_company(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(post_id, production_company_id)
);

CREATE TRIGGER post_production_company_timestamp_update
    BEFORE UPDATE ON post_production_company
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamp();
