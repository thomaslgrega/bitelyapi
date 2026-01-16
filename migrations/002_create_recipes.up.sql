CREATE TABLE IF NOT EXISTS recipes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    instructions TEXT,
    thumbnail_url TEXT,
    calories INTEGER,
    total_cook_time INTEGER,
    created_at TIMESTAMP DEFAULT now()
);