DROP INDEX IF EXISTS recipes_user_id_idx;

ALTER TABLE recipes
DROP COLUMN user_id;
