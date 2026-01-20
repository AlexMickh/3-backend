CREATE TABLE IF NOT EXISTS tokens(
    token TEXT PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) NOT NULL,
    type TEXT CHECK(type IN ('validate-email', 'change-password')),
    expires_at TEXT NOT NULL
);