BEGIN;

ALTER TABLE post
DROP CONSTRAINT post_playlist_db_id_fkey;

ALTER TABLE post
ADD CONSTRAINT post_playlist_db_id_fkey
FOREIGN KEY (playlist_db_id)
REFERENCES playlist(id);

COMMIT;