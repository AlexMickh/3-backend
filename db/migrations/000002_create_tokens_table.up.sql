CREATE TYPE token_type AS enum(
    'validate-email', 
    'change-password'
);

CREATE TABLE IF NOT EXISTS tokens(
    token TEXT PRIMARY KEY,
    user_id UUID REFERENCES users(id) NOT NULL,
    type token_type,
    expires_at TIMESTAMP NOT NULL
);