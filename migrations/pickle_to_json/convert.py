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


if __name__ == "__main__":
    load_dotenv()
    pickle_to_json()
