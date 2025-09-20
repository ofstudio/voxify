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

-- Processes table
CREATE TABLE processes
(
    id                       INTEGER PRIMARY KEY AUTOINCREMENT,
    step                     TEXT     NOT NULL CHECK (step IN ('creating', 'downloading', 'publishing')),
    status                   TEXT     NOT NULL CHECK (status IN ('in_progress', 'success', 'failed')),
    error                    TEXT,
    request_id               TEXT     NOT NULL UNIQUE,
    request_user_id          INTEGER  NOT NULL,
    request_chat_id          INTEGER  NOT NULL,
    request_message_id       INTEGER  NOT NULL,
    request_url              TEXT     NOT NULL,
    request_download_format  TEXT     NOT NULL,
    request_download_quality TEXT     NOT NULL,
    request_force            BOOLEAN  NOT NULL DEFAULT FALSE,
    episode_id               INTEGER,
    created_at               DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at               DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- Foreign key constraint
    FOREIGN KEY (episode_id) REFERENCES episodes (id) ON DELETE SET NULL
);

-- Indexes for better performance
CREATE INDEX idx_episodes_original_url ON episodes (original_url);
CREATE INDEX idx_episodes_created_at ON episodes (created_at);
CREATE INDEX idx_processes_status ON processes (status);
CREATE INDEX idx_processes_step ON processes (step);
CREATE INDEX idx_processes_request_url ON processes (request_url);
CREATE INDEX idx_processes_created_at ON processes (created_at);
CREATE INDEX idx_processes_updated_at ON processes (updated_at);

-- Trigger to automatically update updated_at timestamp
CREATE TRIGGER processes_updated_at
    AFTER UPDATE
    ON processes
    FOR EACH ROW
BEGIN
    UPDATE processes SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
