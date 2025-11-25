CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    short_id TEXT UNIQUE NOT NULL,
    original_url TEXT NOT NULL
);

CREATE INDEX idx_urls_short_id ON urls (short_id);