-- Recipe matching narrows the corpus by trigram similarity against
-- ingredients.name_norm. The btree index from migration 013 cannot serve that:
-- btree answers equality and prefix ranges, and the similarity operators
-- sequential-scan straight past it. A GIN trigram index is what narrowing
-- needs; the btree one can stay, it just serves nothing here.
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX ingredients_name_norm_trgm_idx
  ON ingredients USING gin (name_norm gin_trgm_ops);
