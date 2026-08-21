-- Restores the default that 002 created the column with.
ALTER TABLE recipes
ALTER COLUMN id SET DEFAULT gen_random_uuid();
