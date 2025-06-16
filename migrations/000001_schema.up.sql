-- Used for fuzzy search
CREATE EXTENSION pg_trgm;


CREATE TABLE app_user (
    id SERIAL PRIMARY KEY,
    token VARCHAR(2048),
    google_id VARCHAR(256) UNIQUE,
    facebook_id VARCHAR(256) UNIQUE,
    analytics_id VARCHAR(512),
    name VARCHAR(120),
    email VARCHAR(120),
    picture VARCHAR(512),
    last_seen TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE category (
    id SERIAL PRIMARY KEY,
    name VARCHAR(256) NOT NULL UNIQUE,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    slug VARCHAR(255) NOT NULL UNIQUE
);


CREATE TABLE playlist (
    id SERIAL PRIMARY KEY,
    playlist_id VARCHAR(50) NOT NULL UNIQUE,
    channel_id VARCHAR(50) NOT NULL UNIQUE,
    title VARCHAR(256) NOT NULL,
    channel_title VARCHAR(256),
    thumbnails JSONB NOT NULL,
    channel_thumbnails JSONB NOT NULL,
    description TEXT,
    channel_description TEXT,
    user_id INTEGER REFERENCES app_user(id),
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE post (
    id SERIAL PRIMARY KEY,
    provider VARCHAR(7),
    video_id VARCHAR(20) NOT NULL UNIQUE,
    playlist_id VARCHAR(50),
    title VARCHAR(256) NOT NULL,
    thumbnails JSONB NOT NULL,
    description TEXT,
    short_description TEXT,
    tags TEXT,
    duration VARCHAR(20) NOT NULL,
    upload_date TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    related JSONB,
    user_id INTEGER REFERENCES app_user(id),
    playlist_db_id INTEGER REFERENCES playlist(id),
    search_vector tsvector, -- search vector column
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    category_id INTEGER REFERENCES category(id)
);


-- Create function to automatically update the search_vector column
CREATE FUNCTION update_post_search_vector()
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
CREATE TRIGGER tsvector_update
BEFORE INSERT OR UPDATE ON post
FOR EACH ROW EXECUTE FUNCTION update_post_search_vector();


-- Create GIN index on the post search_vector column
CREATE INDEX idx_post_search_vector ON post USING GIN (search_vector);


-- Create GIN index on the title column for the pg_trgm
CREATE INDEX idx_post_title_trgm ON post USING GIN (title gin_trgm_ops);


CREATE TABLE deleted_post (
    id SERIAL PRIMARY KEY,
    provider VARCHAR(7),
    video_id VARCHAR(20) NOT NULL UNIQUE,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE post_fave (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES app_user(id) ON DELETE CASCADE,
    post_id INTEGER REFERENCES post(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, post_id) -- Prevent duplicates
);


CREATE TABLE post_like (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES app_user(id) ON DELETE CASCADE,
    post_id INTEGER REFERENCES post(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, post_id) -- Prevent duplicates
);


CREATE TABLE page (
    id SERIAL PRIMARY KEY,
    title VARCHAR(256) NOT NULL,
    content TEXT,
    slug VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
