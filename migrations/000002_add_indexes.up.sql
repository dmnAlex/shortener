CREATE UNIQUE INDEX idx_unique_urls_original_url ON urls (original_url);
CREATE INDEX idx_urls_user_id ON urls (user_id);