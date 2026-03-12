ALTER TABLE ingredients ADD COLUMN name_norm TEXT;

UPDATE ingredients
SET name_norm = lower(trim(name));

CREATE INDEX ingredients_name_norm_idx ON ingredients(name_norm);
