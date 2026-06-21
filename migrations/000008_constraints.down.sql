BEGIN;

DROP INDEX idx_post_playlist_composite;

ALTER TABLE post
DROP CONSTRAINT post_playlist_composite_fkey;

ALTER TABLE post
ADD CONSTRAINT post_playlist_db_id_fkey
FOREIGN KEY (playlist_db_id) REFERENCES playlist(id) ON DELETE CASCADE;

ALTER TABLE playlist
DROP CONSTRAINT playlist_id_pair_unique;

COMMIT;