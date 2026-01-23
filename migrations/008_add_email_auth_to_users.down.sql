DROP INDEX IF EXISTS users_email_unique;

ALTER TABLE users
DROP COLUMN password_hash;

ALTER TABLE users
ALTER COLUMN apple_sub SET NOT NULL;