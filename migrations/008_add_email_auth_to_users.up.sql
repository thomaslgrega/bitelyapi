ALTER TABLE users
ALTER COLUMN apple_sub DROP NOT NULL;

ALTER TABLE users
ADD COLUMN password_hash TEXT;

CREATE UNIQUE INDEX users_email_unique
ON users(email)
WHERE email IS NOT NULL;