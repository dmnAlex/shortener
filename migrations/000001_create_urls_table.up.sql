CREATE TABLE IF NOT EXISTS urls (
    short_id TEXT PRIMARY KEY,
    original_url TEXT NOT NULL,
    user_id TEXT NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);