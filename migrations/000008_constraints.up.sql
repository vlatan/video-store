BEGIN;

ALTER TABLE playlist
ADD CONSTRAINT playlist_id_pair_unique UNIQUE (id, playlist_id);

ALTER TABLE post
DROP CONSTRAINT post_playlist_db_id_fkey;

ALTER TABLE post
ADD CONSTRAINT post_playlist_composite_fkey
FOREIGN KEY (playlist_db_id, playlist_id)
REFERENCES playlist (id, playlist_id)
MATCH FULL
ON DELETE CASCADE;

CREATE INDEX idx_post_playlist_composite ON post (playlist_db_id, playlist_id);

COMMIT;