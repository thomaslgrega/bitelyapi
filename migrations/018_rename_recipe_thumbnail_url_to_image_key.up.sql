ALTER TABLE recipes RENAME COLUMN thumbnail_url TO image_key;

-- The column holds an object key, not a URL: the bucket's public hostname is
-- configuration, so a Recipe row never names it (ADR-0006). Existing values are
-- absolute picsum URLs from the dev seed and are not keys.
UPDATE recipes SET image_key = '';

-- No image is the empty string rather than NULL, so every read scans into a
-- plain string the way the other columns already do.
ALTER TABLE recipes ALTER COLUMN image_key SET DEFAULT '';
ALTER TABLE recipes ALTER COLUMN image_key SET NOT NULL;
