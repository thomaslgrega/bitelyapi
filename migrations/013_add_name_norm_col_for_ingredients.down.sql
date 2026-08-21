-- The up migration's backfill needs no inverse: it goes with the column.
DROP INDEX IF EXISTS ingredients_name_norm_idx;

ALTER TABLE ingredients
DROP COLUMN name_norm;
