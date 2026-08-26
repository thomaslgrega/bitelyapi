# A Name Query matches Recipe names only, and matches them fuzzily

`GET /recipes?name=` answers a Name Query with Shared Recipes whose **name** the query reaches, ordered by how closely it reaches them. Two decisions the endpoint had to settle, and the reasons they went the way they did.

**Names only, not Ingredients.** A query of "chicken" will be typed by people meaning both things, but the ingredient half already has an endpoint: `POST /recipes/match` answers a whole pantry with Coverage, matched Ingredients and missing ones. Folding Ingredients into the Name Query would give the same question two answers with different shapes, and would make a Recipe that merely calls for chicken indistinguishable in the response from one named for it. Names-only is the line that keeps the two features nameable — a Name Query reaches a Recipe the user already has in mind; a Match offers Recipes they do not.

**Fuzzy, via trigram word similarity.** Recipe names are short, foreign, and misspelled by exactly the people searching for them — shakshuka, focaccia, ratatouille. Prefix or substring matching answers those with nothing. So the match is `pg_trgm` word similarity against a generated `recipes.name_norm`, the same mechanism ADR-0002 chose for Ingredients, with a GIN trigram index behind it.

The threshold is **0.5**, not the 0.3 that ingredient matching uses, because the two are doing opposite jobs. Ingredient narrowing is deliberately generous: the matching package re-scores every row it returns, so recall there is free. Nothing re-scores a Name Query — whatever clears the threshold is the answer the user reads. Measured against a sample corpus, 0.3 is where names that merely share a stem arrive: "banana" returns Chana Masala, "chicken" returns Chocolate Chip Cookies. 0.5 drops those while still clearing every misspelling worth serving (shakshouka/shakshuka 0.615, focacia/focaccia 0.700, ratatouile/ratatouille 0.818).

Word similarity scores the query against the best-matching run of words in the name rather than against the whole name, which is what lets a half-typed query work — "shak" reaches "Green Shakshuka" at 0.8 — so no separate prefix or substring branch sits beside it. An earlier draft had one; measured, it added only matches from *inside* a word ("toui" reaching "Ratatouille") along with noise from two-character queries, so it was removed.

## Consequences

Search is as loose as matching is, and for the same reason ADR-0002 records: "chicken" also returns Chickpea Curry, which scores 0.625 and cannot be separated from a real misspelling by any threshold. Ordering is the mitigation — an exact name scores 1.0 and leads — and the catalog ADR-0002 names as its destination would fix search too.

Ordering is by similarity alone, with name and id only as tiebreakers. There is no popularity or recency signal in the corpus to blend in yet; when there is one, this is the ordering it changes.

The answer is capped at 50 with no pagination, matching `POST /recipes/match`. Past 50 the trigram tail is noise, and a Name Query is how someone reaches a Recipe they already have in mind, not how they browse.

The parameter is spelled `name`, not the `q` the issue sketched. `q` is the web-wide convention for free-text search, but it promises a search over everything, which is the reading this ADR exists to refuse — and the endpoint's only other parameter, `category`, is spelled out. `?name=` says what is matched, and stops being the right name on the day the search stops being names-only, which is the correct time to revisit it.

`category` composes with `name`, narrowing the search rather than widening it. Neither parameter is required on its own, but a request with neither is a 400 rather than the whole corpus.
