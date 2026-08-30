ALTER TABLE recipes ALTER COLUMN image_key DROP NOT NULL;
ALTER TABLE recipes ALTER COLUMN image_key DROP DEFAULT;

UPDATE recipes SET image_key = NULL WHERE image_key = '';

ALTER TABLE recipes RENAME COLUMN image_key TO thumbnail_url;
