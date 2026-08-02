-- NOTE: original post_review.id values are not recoverable (dropped in the up migration).
-- This restores the shape of the table, not the original data verbatim.

ALTER TABLE post_review
ALTER COLUMN title DROP NOT NULL,
ALTER COLUMN review DROP NOT NULL;

ALTER TABLE post_review DROP CONSTRAINT post_review_pkey;

ALTER TABLE post_review ADD COLUMN id SERIAL;
ALTER TABLE post_review ADD COLUMN user_id INTEGER;
ALTER TABLE post_review ADD COLUMN post_id INTEGER;

UPDATE post_review AS pr
SET user_id = prat.user_id, post_id = prat.post_id
FROM post_rating AS prat
WHERE pr.rating_id = prat.id;

ALTER TABLE post_review ADD PRIMARY KEY (id);
ALTER TABLE post_review
ADD CONSTRAINT post_review_user_id_fkey FOREIGN KEY (user_id) REFERENCES app_user(id) ON DELETE CASCADE,
ADD CONSTRAINT post_review_post_id_fkey FOREIGN KEY (post_id) REFERENCES post(id) ON DELETE CASCADE,
ADD CONSTRAINT post_review_user_id_post_id_key UNIQUE (user_id, post_id);
CREATE INDEX idx_post_review_post_id ON post_review(post_id);

ALTER TABLE post_review DROP COLUMN rating_id;

ALTER TABLE post_rating
ALTER COLUMN user_id DROP NOT NULL,
ALTER COLUMN post_id DROP NOT NULL,
ALTER COLUMN rating DROP NOT NULL;

ALTER TABLE post_like
ALTER COLUMN user_id DROP NOT NULL,
ALTER COLUMN post_id DROP NOT NULL;

ALTER TABLE post_fave
ALTER COLUMN user_id DROP NOT NULL,
ALTER COLUMN post_id DROP NOT NULL;
