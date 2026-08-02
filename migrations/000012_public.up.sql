-- Make the public id unique
ALTER TABLE app_user ADD CONSTRAINT app_user_public_id_key UNIQUE (public_id);
