# Trigram token matching instead of an ingredient catalog

Matching Pantry Items against Ingredient Terms needs to survive real phrasing: "tomato" against "tomatoes", "chicken breast" against "boneless skinless chicken breasts". The correct model is a curated catalog of canonical ingredients with aliases, which recipe lines resolve to when written. We deferred it, because it requires a catalog we do not have and a backfill of every Ingredient already stored, and shipped fuzzy string matching instead: `pg_trgm` trigram similarity in Postgres, the same algorithm ported to Swift for the local pass.

Similarity is compared **per token, not per whole string**. Whole-string comparison fails the motivating case — the extra descriptors in "boneless skinless chicken breasts" flood the trigram union and drive similarity below any usable threshold. So both sides split into words, drop descriptor stopwords, and treat a Pantry Item as matching an Ingredient Term when any token pair scores at or above 0.3. A match is binary: Coverage stays a clean integer ratio so the interface can say "you have 4 of 6 ingredients", which fractional credit would make unsayable.

## Consequences

Token matching is deliberately loose and will produce false matches — "chicken breast" also matches "chicken stock". Precision is what we traded for shipping without a catalog, and the catalog is the known destination that buys it back. The glossary's split between an Ingredient, which is a line on one Recipe, and an Ingredient Term, which is the shared normalized food name, exists so the catalog can be introduced later as a resolution step behind Ingredient Term rather than as a schema rewrite.

Staples are defined in the glossary but not yet excluded from Coverage, because identifying them reliably needs the catalog. Until then there is no cap on missing ingredients: a cap combined with unrecognized Staples would hide good Matches behind "missing salt, missing pepper", so Coverage ordering alone buries weak Matches. A hardcoded Staples list was rejected as a stopgap because it would create a second source of truth, in two languages, to reconcile later.
