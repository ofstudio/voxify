-- Episodes table
CREATE TABLE episodes
(
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    title          TEXT     NOT NULL,
    description    TEXT,
    thumbnail_file TEXT,
    media_file     TEXT     NOT NULL,
    media_type     TEXT     NOT NULL,
    media_duration INTEGER  NOT NULL DEFAULT 0,
    media_size     INTEGER  NOT NULL DEFAULT 0,
    author         TEXT     NOT NULL,
    original_url   TEXT     NOT NULL,
    canonical_url  TEXT     NOT NULL,
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for better performance
CREATE INDEX idx_episodes_original_url ON episodes (original_url);
CREATE INDEX idx_episodes_canonical_url ON episodes (canonical_url);
CREATE INDEX idx_episodes_created_at ON episodes (created_at);
