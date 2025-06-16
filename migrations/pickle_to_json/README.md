# DATABASE Improvements

We need the production database data and the schema:
1. Dump the production data with the schema.
2. Bring down the docker compose.
3. Comment out the `./migrations:/docker-entrypoint-initdb.d` volume in `compose.yaml`.
4. Delete the postgress data volume.
5. Run the docker compose.
5. Copy the database dump into the docker container.
7. Import/restore the data with its original schema. If you see ONLY errors that complain that the role `postgres` does not exist you are good to go.


This is starting point, we have the exact same database from the production. The following are steps for manipulating and adapting the data to be able to fit into the new, improved databse schema.


## Preliminary Clean Up

Rename the table "user" which is a reserved postgres keyword.
``` sql
ALTER TABLE "user" RENAME TO app_user;
ALTER SEQUENCE public.user_id_seq RENAME TO app_user_id_seq;
```

Drop the alembic table.
``` sql
DROP TABLE alembic_version;
```


## Pickle to JSON Conversion

Because `pickle` is a Python specific format and we're rewriting this app in Go, Go can't actually read the pickled `BYTEA` database columns/values so we need to convert those to `JSONB`. But because Postgres can't do this we need a Python script for conversion.

Add these columns to `playlist` and `post` tables:
``` sql
ALTER TABLE playlist ADD COLUMN thumbnails_json JSONB;
ALTER TABLE playlist ADD COLUMN channel_thumbnails_json JSONB;
ALTER TABLE post ADD COLUMN thumbnails_json JSONB;
ALTER TABLE post ADD COLUMN related_json JSONB;
```

Change the `DB_HOST` in the `.env` file temporarely to `localhost`. Run the `convert.py` script which will extract the pickled content from each row, convert it to JSON and place it in these new columns.


Switch the column names:
``` sql
ALTER TABLE playlist RENAME COLUMN thumbnails TO thumbnails_pickle;
ALTER TABLE playlist RENAME COLUMN thumbnails_json TO thumbnails;
ALTER TABLE playlist RENAME COLUMN channel_thumbnails TO channel_thumbnails_pickle;
ALTER TABLE playlist RENAME COLUMN channel_thumbnails_json TO channel_thumbnails;
ALTER TABLE post RENAME COLUMN thumbnails TO thumbnails_pickle;
ALTER TABLE post RENAME COLUMN thumbnails_json TO thumbnails;
ALTER TABLE post RENAME COLUMN "similar" TO related_pickle;
ALTER TABLE post RENAME COLUMN related_json TO related;
```

Set NOT NULL constraint to the newly renamed columns:
``` sql
ALTER TABLE playlist ALTER COLUMN thumbnails SET NOT NULL;
ALTER TABLE playlist ALTER COLUMN channel_thumbnails SET NOT NULL;
ALTER TABLE post ALTER COLUMN thumbnails SET NOT NULL;
```

If everything is OK, drop the pickled columns:
```sql
ALTER TABLE playlist DROP COLUMN thumbnails_pickle;
ALTER TABLE playlist DROP COLUMN channel_thumbnails_pickle;
ALTER TABLE post DROP COLUMN thumbnails_pickle;
ALTER TABLE post DROP COLUMN related_pickle;
```

## Remove Orphaned and Duplicate Data

Now move on to deleting orphaned and duplicate data from the `post_fave` and `post_like` tables and setting up constraints so in future we don't end up with orphaned and/or duplicate data.

Clean up orphaned faves and likes (foreign key violations)
``` sql
-- Remove post_fave records that reference non-existent posts
DELETE FROM post_fave 
WHERE post_id NOT IN (SELECT id FROM post);

-- Remove post_fave records that reference non-existent users
DELETE FROM post_fave 
WHERE user_id NOT IN (SELECT id FROM app_user);

-- Remove post_like records that reference non-existent posts
DELETE FROM post_like 
WHERE post_id NOT IN (SELECT id FROM post);

-- Remove post_like records that reference non-existent users
DELETE FROM post_like 
WHERE user_id NOT IN (SELECT id FROM app_user);
```

Check for duplicates before removing them.  
First review what will be deleted if anything.
``` sql
SELECT user_id, post_id, COUNT(*) as duplicate_count
FROM post_fave 
GROUP BY user_id, post_id 
HAVING COUNT(*) > 1
ORDER BY duplicate_count DESC;

SELECT user_id, post_id, COUNT(*) as duplicate_count
FROM post_like 
GROUP BY user_id, post_id 
HAVING COUNT(*) > 1
ORDER BY duplicate_count DESC;
```

Remove duplicates, keeping the earliest record:
``` sql
-- For post_fave duplicates
DELETE FROM post_fave AS a USING (
    SELECT MIN(id) as min_id, user_id, post_id
    FROM post_fave 
    GROUP BY user_id, post_id
    HAVING COUNT(*) > 1
) AS b
WHERE a.user_id = b.user_id 
  AND a.post_id = b.post_id 
  AND a.id > b.min_id;

-- For post_like duplicates
DELETE FROM post_like AS a USING (
    SELECT MIN(id) as min_id, user_id, post_id
    FROM post_like 
    GROUP BY user_id, post_id
    HAVING COUNT(*) > 1
) AS b
WHERE a.user_id = b.user_id 
  AND a.post_id = b.post_id 
  AND a.id > b.min_id;
```

Verify cleanup was successful:
``` sql
-- These should return 0 rows
SELECT user_id, post_id, COUNT(*) 
FROM post_fave 
GROUP BY user_id, post_id 
HAVING COUNT(*) > 1;

SELECT user_id, post_id, COUNT(*) 
FROM post_like 
GROUP BY user_id, post_id 
HAVING COUNT(*) > 1;
```

Now apply the schema changes
``` sql
-- Add the CASCADE options and UNIQUE constraints
ALTER TABLE post_fave 
  DROP CONSTRAINT post_fave_user_id_fkey,
  DROP CONSTRAINT post_fave_post_id_fkey,
  ADD CONSTRAINT post_fave_user_id_fkey 
    FOREIGN KEY (user_id) REFERENCES app_user(id) ON DELETE CASCADE,
  ADD CONSTRAINT post_fave_post_id_fkey 
    FOREIGN KEY (post_id) REFERENCES post(id) ON DELETE CASCADE,
  ADD CONSTRAINT post_fave_user_post_unique 
    UNIQUE(user_id, post_id);

ALTER TABLE post_like 
  DROP CONSTRAINT post_like_user_id_fkey,
  DROP CONSTRAINT post_like_post_id_fkey,
  ADD CONSTRAINT post_like_user_id_fkey 
    FOREIGN KEY (user_id) REFERENCES app_user(id) ON DELETE CASCADE,
  ADD CONSTRAINT post_like_post_id_fkey 
    FOREIGN KEY (post_id) REFERENCES post(id) ON DELETE CASCADE,
  ADD CONSTRAINT post_like_user_post_unique 
    UNIQUE(user_id, post_id);
```


## Enabling Full Text Search

We used redis in the python version of this app, but we'll switch to postgres itself. It has powerfull capabilities.


``` sql
-- Used for fuzzy search
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
```

Run the Golang up, it should work.

Now we need to:
1. Dump JUST the data we manipulated.
2. Bring down the docker compose.
3. Delete the postgres data volume.
4. Uncomment the `./migrations:/docker-entrypoint-initdb.d` volume.
5. Run the docker compose.
6. Copy the database dump into the docker container.
7. Import/restore JUST the data.

If everything is okay dump the data with the schema and restore it to a new database in production.
