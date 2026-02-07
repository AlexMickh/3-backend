CREATE TABLE IF NOT EXISTS products(
    id INTEGER PRIMARY KEY,
    category_id INTEGER REFERENCES categories(id) ON DELETE CASCADE NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    price INTEGER CHECK (price > 0) NOT NULL, -- stores kopeck
    quantity INTEGER CHECK (quantity >= 0) NOT NULL,
    existing_sizes TEXT NOT NULL,
    image_ulr TEXT NOT NULL,
    pieces_sold INTEGER DEFAULT 0,
    discount INTEGER CHECK (discount >= 0 AND discount < price) DEFAULT 0,
    discount_expires_at TEXT NULL,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT
);