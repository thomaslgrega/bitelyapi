-- Restore name_norm to a plain writable column, backfilled as migration 013 did.
DROP INDEX ingredients_name_norm_idx;

ALTER TABLE ingredients
DROP COLUMN name_norm;

ALTER TABLE ingredients
ADD COLUMN name_norm TEXT;

UPDATE ingredients
SET name_norm = lower(btrim(name, E' \t\r\n'));

CREATE INDEX ingredients_name_norm_idx ON ingredients(name_norm);
