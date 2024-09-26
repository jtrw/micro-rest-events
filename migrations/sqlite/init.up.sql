CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL,
    user_id TEXT NULL,
    type TEXT NULL,
    status TEXT NULL,
    caption TEXT NULL,
    message TEXT NULL,
    is_seen BOOLEAN DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX events_user_id_idx ON events (user_id);

CREATE INDEX events_status_index ON events (status);

CREATE INDEX events_user_id_status_index ON events (user_id, status);
