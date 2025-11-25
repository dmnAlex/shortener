CREATE TABLE IF NOT EXISTS users (
    user_id TEXT NOT NULL,
    short_id TEXT NOT NULL,
    PRIMARY KEY (user_id, short_id)
);