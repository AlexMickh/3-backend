ALTER TABLE users ADD COLUMN is_email_verified BOOLEAN DEFAULT FALSE;

CREATE TYPE token_type AS ENUM(
    'validate-email',
    'change-password'
);

CREATE TABLE IF NOT EXISTS tokens(
    token TEXT PRIMARY KEY,
    user_id UUID REFERENCES users(id) NOT NULL,
    type token_type NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);