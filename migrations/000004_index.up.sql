-- Create index needed for the cursor infinite scroll
CREATE INDEX idx_post_upload_date_id ON post (upload_date DESC, id DESC);