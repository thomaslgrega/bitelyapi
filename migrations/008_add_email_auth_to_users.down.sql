-- apple_sub is not restored to NOT NULL. The up migration dropped that
-- constraint so users could sign up with an email instead, and every such row
-- has a null apple_sub, so restoring it would fail against any real data.
DROP INDEX IF EXISTS users_email_unique;

ALTER TABLE users
DROP COLUMN password_hash;
