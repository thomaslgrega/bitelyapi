-- A Name Query matches Recipe names fuzzily (ADR-0004), which needs the same
-- two things ingredient narrowing needed: a normalized column to compare
-- against, and a GIN trigram index so the similarity operator does not
-- sequential-scan the corpus.
-- The generation expression matches ingredients.name_norm exactly, including
-- the trim charset, so the two columns normalize a name the same way.
-- Adding a stored generated column backfills every row, which rewrites the
-- table under an ACCESS EXCLUSIVE lock.
ALTER TABLE recipes
ADD COLUMN name_norm TEXT GENERATED ALWAYS AS (lower(btrim(name, E' \t\r\n'))) STORED;

CREATE INDEX recipes_name_norm_trgm_idx
  ON recipes USING gin (name_norm gin_trgm_ops);
