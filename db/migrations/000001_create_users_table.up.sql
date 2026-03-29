CREATE TYPE user_role AS ENUM(
    'user',
    'admin'
);

CREATE TABLE IF NOT EXISTS users(
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    email VARCHAR(100) UNIQUE,
    phone VARCHAR(12),
    password TEXT NOT NULL,
    role user_role DEFAULT 'user',
    is_email_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);
