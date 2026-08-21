-- The extension can only go once no column is typed citext.
ALTER TABLE users
ALTER COLUMN email TYPE TEXT;

DROP EXTENSION IF EXISTS citext;
