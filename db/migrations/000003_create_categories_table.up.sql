CREATE TABLE IF NOT EXISTS categories(
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE,
    created_at TEXT DEFAULT (datetime('now'))
);