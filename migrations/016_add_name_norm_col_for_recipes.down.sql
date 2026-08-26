DROP INDEX IF EXISTS recipes_name_norm_trgm_idx;

ALTER TABLE recipes
DROP COLUMN name_norm;
