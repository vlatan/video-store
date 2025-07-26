import os
import pickle
import time
import json
import psycopg
from psycopg.conninfo import make_conninfo
from dotenv import load_dotenv


def pickle_to_json():
    conninfo = make_conninfo(
        user=os.getenv("DB_USERNAME"),
        password=os.getenv("DB_PASSWORD"),
        host=os.getenv("DB_HOST"),
        port=os.getenv("DB_PORT"),
        dbname=os.getenv("DB_DATABASE"),
    )

    print("Script started...")
    start = time.perf_counter()

    # Connect to an existing database
    with psycopg.connect(conninfo) as conn:
        # Open a cursor to perform database operations
        with conn.cursor() as cur:

            # Rename the table "user" which is a reserved postgres keyword
            cur.execute(
                """
                ALTER TABLE "user" RENAME TO app_user;
                ALTER SEQUENCE public.user_id_seq RENAME TO app_user_id_seq;
                """
            )

            # Drop the alembic table
            cur.execute(
                """
                DROP TABLE alembic_version;
                """
            )

            # Add these columns to playlist and post tables
            cur.execute(
                """
                ALTER TABLE playlist ADD COLUMN thumbnails_json JSONB;
                ALTER TABLE playlist ADD COLUMN channel_thumbnails_json JSONB;
                ALTER TABLE post ADD COLUMN thumbnails_json JSONB;
                ALTER TABLE post ADD COLUMN related_json JSONB;
                """
            )

            # Fix playlist thumbnails
            cur.execute(
                """
                SELECT id, thumbnails FROM playlist
                WHERE thumbnails IS NOT NULL
                """
            )

            for row_id, pickled_data in cur.fetchall():
                data = pickle.loads(pickled_data)
                json_data = json.dumps(data)
                cur.execute(
                    "UPDATE playlist SET thumbnails_json = %s WHERE id = %s",
                    (json_data, row_id),
                )

            # Fix channel thumbnails
            cur.execute(
                """SELECT id, channel_thumbnails FROM playlist
                WHERE channel_thumbnails IS NOT NULL"""
            )

            for row_id, pickled_data in cur.fetchall():
                data = pickle.loads(pickled_data)
                json_data = json.dumps(data)
                cur.execute(
                    "UPDATE playlist SET channel_thumbnails_json = %s WHERE id = %s",
                    (json_data, row_id),
                )

            # Fix post thumbnails
            cur.execute(
                """
                SELECT id, thumbnails FROM post
                WHERE thumbnails IS NOT NULL
                """
            )

            for row_id, pickled_data in cur.fetchall():
                data = pickle.loads(pickled_data)
                json_data = json.dumps(data)
                cur.execute(
                    "UPDATE post SET thumbnails_json = %s WHERE id = %s",
                    (json_data, row_id),
                )

            # Fix similar in post
            cur.execute(
                """
                SELECT id, 'similar' FROM post
                WHERE 'similar' IS NOT NULL
                """
            )

            for row_id, data in cur.fetchall():
                json_data = json.dumps(data)
                cur.execute(
                    "UPDATE post SET related_json = %s WHERE id = %s",
                    (json_data, row_id),
                )

            # Swap the columns
            cur.execute(
                """
                -- Switch the column names
                ALTER TABLE playlist RENAME COLUMN thumbnails TO thumbnails_pickle;
                ALTER TABLE playlist RENAME COLUMN thumbnails_json TO thumbnails;
                ALTER TABLE playlist RENAME COLUMN channel_thumbnails TO channel_thumbnails_pickle;
                ALTER TABLE playlist RENAME COLUMN channel_thumbnails_json TO channel_thumbnails;
                ALTER TABLE post RENAME COLUMN thumbnails TO thumbnails_pickle;
                ALTER TABLE post RENAME COLUMN thumbnails_json TO thumbnails;
                ALTER TABLE post RENAME COLUMN "similar" TO related_pickle;
                ALTER TABLE post RENAME COLUMN related_json TO related;

                -- Set NOT NULL constraint to the newly renamed columns:
                ALTER TABLE playlist ALTER COLUMN thumbnails SET NOT NULL;
                ALTER TABLE playlist ALTER COLUMN channel_thumbnails SET NOT NULL;
                ALTER TABLE post ALTER COLUMN thumbnails SET NOT NULL;

                -- Drop the pickled columns
                ALTER TABLE playlist DROP COLUMN thumbnails_pickle;
                ALTER TABLE playlist DROP COLUMN channel_thumbnails_pickle;
                ALTER TABLE post DROP COLUMN thumbnails_pickle;
                ALTER TABLE post DROP COLUMN related_pickle;
                """
            )

            # Clean up orphaned faves and likes (foreign key violations)
            cur.execute(
                """
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
                """
            )

            # Remove duplicate likes/faves, keeping the earliest record
            cur.execute(
                """
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
                """
            )

            # apply schema changes to likes, faves
            cur.execute(
                """
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
                """
            )

            # Enable Full Text Search
            cur.execute(
                """
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
                """
            )

        conn.commit()

        print("Script done in: ", time.perf_counter() - start)


if __name__ == "__main__":
    load_dotenv()
    pickle_to_json()
