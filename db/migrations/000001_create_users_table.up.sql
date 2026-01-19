CREATE TYPE user_role AS ENUM(
    'user',
    'admin'
); 

CREATE TABLE IF NOT EXISTS users(
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    email TEXT UNIQUE,
    phone VARCHAR(12) UNIQUE,
    password TEXT,
    role user_role DEFAULT 'user',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ
);