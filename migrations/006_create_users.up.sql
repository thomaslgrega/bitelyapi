CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    apple_sub TEXT NOT NULL UNIQUE,
    email TEXT,
    created_at TIMESTAMP DEFAULT now()
);
