CREATE TABLE IF NOT EXISTS users(
    id INTEGER PRIMARY KEY,
    email TEXT UNIQUE,
    phone TEXT CHECK(length(phone) <= 12),
    password TEXT,
    role TEXT DEFAULT 'user' CHECK(role IN ('user', 'admin')),
    is_email_verified INTEGER DEFAULT 0 CHECK(is_email_verified IN (0, 1)),
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT
);
