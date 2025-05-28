import os
import pickle
import time
import json
import psycopg
from psycopg.conninfo import make_conninfo
from dotenv import load_dotenv

"""
-- Add these columns to playlist and post tables
ALTER TABLE playlist ADD COLUMN thumbnails_json JSONB;
ALTER TABLE playlist ADD COLUMN channel_thumbnails_json JSONB;
ALTER TABLE post ADD COLUMN thumbnails_json JSONB;
ALTER TABLE post ADD COLUMN related_json JSONB;

-- RUN THIS PYTHON SCRIPT TO COPY AND CONVERT DATA
-- FROM THE 'BYTEA' COLUMNS TO THE NEW JSONB COLUMNS
"""


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

        conn.commit()

        print("Script done in: ", time.perf_counter() - start)


"""
-- Switch the columns names
ALTER TABLE playlist RENAME COLUMN thumbnails TO thumbnails_pickle;
ALTER TABLE playlist RENAME COLUMN thumbnails_json TO thumbnails;
ALTER TABLE playlist RENAME COLUMN channel_thumbnails TO channel_thumbnails_pickle;
ALTER TABLE playlist RENAME COLUMN channel_thumbnails_json TO channel_thumbnails;
ALTER TABLE post RENAME COLUMN thumbnails TO thumbnails_pickle;
ALTER TABLE post RENAME COLUMN thumbnails_json TO thumbnails;
ALTER TABLE post RENAME COLUMN "similar" TO related_pickle;
ALTER TABLE post RENAME COLUMN related_json TO related;

-- Set NOT NULL to the newly renamed columns
ALTER TABLE playlist ALTER COLUMN thumbnails SET NOT NULL;
ALTER TABLE playlist ALTER COLUMN channel_thumbnails SET NOT NULL;
ALTER TABLE post ALTER COLUMN thumbnails SET NOT NULL;

--Drop the BYTEA columns if everything is OK
ALTER TABLE playlist DROP COLUMN thumbnails_pickle;
ALTER TABLE playlist DROP COLUMN channel_thumbnails_pickle;
ALTER TABLE post DROP COLUMN thumbnails_pickle;
ALTER TABLE post DROP COLUMN related_pickle;

-- FINALLY CHANGE FROM BYTEA TO JSONB IN THE SCHEMA
-- AND CHANGE COLUMN NAME "similar" TO related IN post
"""


if __name__ == "__main__":
    load_dotenv()
    pickle_to_json()
