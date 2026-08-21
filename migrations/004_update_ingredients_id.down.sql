-- Restores the default that 003 created the column with.
ALTER TABLE ingredients
ALTER COLUMN id SET DEFAULT gen_random_uuid();
