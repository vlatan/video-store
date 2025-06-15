-- Make user and post ID combination unique on user like or fave
ALTER TABLE post_like ADD CONSTRAINT unique_user_post_like UNIQUE (user_id, post_id);
ALTER TABLE post_fave ADD CONSTRAINT unique_user_post_fave UNIQUE (user_id, post_id);