-- post_like / post_fave: enforce NOT NULL on FK columns
DELETE FROM post_like WHERE user_id IS NULL OR post_id IS NULL;
ALTER TABLE post_like
ALTER COLUMN user_id SET NOT NULL,
ALTER COLUMN post_id SET NOT NULL;

DELETE FROM post_fave WHERE user_id IS NULL OR post_id IS NULL;
ALTER TABLE post_fave
ALTER COLUMN user_id SET NOT NULL,
ALTER COLUMN post_id SET NOT NULL;

-- post_rating: enforce NOT NULL on core columns
DELETE FROM post_rating WHERE user_id IS NULL OR post_id IS NULL OR rating IS NULL;
ALTER TABLE post_rating
ALTER COLUMN user_id SET NOT NULL,
ALTER COLUMN post_id SET NOT NULL,
ALTER COLUMN rating SET NOT NULL;

-- post_review: make it an optional extension of post_rating via shared PK
ALTER TABLE post_review ADD COLUMN rating_id INTEGER REFERENCES post_rating(id) ON DELETE CASCADE;

UPDATE post_review AS pr
SET rating_id = prat.id
FROM post_rating AS prat
WHERE pr.user_id = prat.user_id AND pr.post_id = prat.post_id;

DELETE FROM post_review WHERE rating_id IS NULL OR title IS NULL OR review IS NULL;

-- Dropping these columns automatically drops post_review_pkey, both FK constraints,
-- the (user_id, post_id) unique constraint, and idx_post_review_post_id along with them
-- nothing else references them, so no explicit DROP needed.
ALTER TABLE post_review DROP COLUMN id;
ALTER TABLE post_review DROP COLUMN user_id;
ALTER TABLE post_review DROP COLUMN post_id;

ALTER TABLE post_review ADD PRIMARY KEY (rating_id);
ALTER TABLE post_review
ALTER COLUMN title SET NOT NULL,
ALTER COLUMN review SET NOT NULL;
