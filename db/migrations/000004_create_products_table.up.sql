CREATE TABLE IF NOT EXISTS products(
    id UUID PRIMARY KEY,
    category_id UUID REFERENCES categories(id) ON DELETE CASCADE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    price INTEGER CHECK (price > 0) NOT NULL, -- stores kopeck
    quantity INTEGER CHECK (quantity >= 0) NOT NULL,
    existing_sizes TEXT[] NOT NULL,
    image_url TEXT,
    pieces_sold INTEGER DEFAULT 0,
    discount INTEGER CHECK (discount >= 0 AND discount < 100) DEFAULT 0,
    discount_expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);