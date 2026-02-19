CREATE TABLE IF NOT EXISTS memories (
    id          TEXT PRIMARY KEY,
    type        TEXT NOT NULL,
    content     TEXT NOT NULL,
    source      TEXT NOT NULL,
    importance  REAL NOT NULL DEFAULT 0.5,
    created_at  INTEGER NOT NULL,
    accessed_at INTEGER NOT NULL,
    metadata    TEXT NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS facts (
    key         TEXT PRIMARY KEY,
    value       TEXT NOT NULL,
    confidence  REAL NOT NULL DEFAULT 1.0,
    updated_at  INTEGER NOT NULL
);
