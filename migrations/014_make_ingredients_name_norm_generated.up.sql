-- name_norm is derived from name, so let Postgres maintain it instead of every
-- writer (Go, dev seed, hand-written SQL) remembering to set it.
-- A generation expression cannot be attached to an existing column, so the
-- column is dropped and re-added; adding it backfills every row, which rewrites
-- the table under an ACCESS EXCLUSIVE lock.
-- The trim charset matches Go's strings.TrimSpace, which previously produced
-- this value, rather than trim()'s space-only default.
DROP INDEX ingredients_name_norm_idx;

ALTER TABLE ingredients
DROP COLUMN name_norm;

ALTER TABLE ingredients
ADD COLUMN name_norm TEXT GENERATED ALWAYS AS (lower(btrim(name, E' \t\r\n'))) STORED;

CREATE INDEX ingredients_name_norm_idx ON ingredients(name_norm);
