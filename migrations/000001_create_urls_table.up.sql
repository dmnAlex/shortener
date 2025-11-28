CREATE TABLE IF NOT EXISTS urls (
    short_id TEXT PRIMARY KEY,
    original_url TEXT NOT NULL
);

CREATE INDEX idx_urls_short_id ON urls (short_id);