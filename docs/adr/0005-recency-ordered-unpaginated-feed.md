# The Feed is a bare `GET /recipes`, ordered by recency and unpaginated

`GET /recipes` with neither a category nor a Name Query answers the Feed: a bounded selection of Shared Recipes for a client that has nothing to narrow by yet. Before this it answered 400. The client's workaround was to fan out across all ten categories on every cold launch and merge the responses — ten large responses to fill one grid, and worse as the corpus grows.

**A bare `GET /recipes`, not a new path.** The Feed returns the same `RecipeSummary` list the category and Name Query branches already return, so a client swaps its data source and nothing downstream changes. `/recipes/feed` would have been a second URL for the same resource under a narrower name — and browse-with-no-narrowing is what a bare collection GET means everywhere else. What is lost is the 400 that used to catch a client that forgot its `category`; that request now costs a query instead of an error, which is the trade a browsable collection makes.

**Ordered by recency of sharing.** Bitely measures nothing else. There is no view count, no save count, no rating — a "popular" or "featured" ordering would be an ordering the data cannot support, dressed up as one it can. Newest-first is true, is stable between requests, and rewards sharing. It is also the ordering the client already fakes locally, which is the tell that it is the honest one. The column is `recipes.created_at`, which is the moment of sharing rather than a proxy for it: a row exists in the corpus only because it was shared. It is nullable, so the ordering is `NULLS LAST` — a row with no time on it is the least recent thing in the corpus, not the most. `id DESC` breaks ties so two Recipes shared in the same instant do not swap places between requests.

**No cursor.** The issue raised pagination as an open question; the answer for now is a cap, matching `POST /recipes/match` and the Name Query. A cursor needs the client to be able to name where it stopped, and `RecipeSummary` carries no `created_at` — so a cursor means either an envelope around the list or a header, and the envelope is exactly the response-shape change the issue asked to avoid. Against a corpus of a few hundred Shared Recipes, fifty is more Feed than the grid shows before someone shares again. When the corpus outgrows that, keyset pagination on `(created_at, id)` is what this ordering was chosen to support, and it arrives with the shape change it needs.

**`limit` lowers a cap; it never raises or invents one.** The Feed's cap is 50, the same fifty the Name Query and `POST /recipes/match` stop at, and absent a `limit` that is what a request gets. Present, `limit` must be a positive whole number — `limit=0`, `limit=2.5` and `limit=` are client bugs and answer 400 rather than quietly becoming some page size nobody asked for — and above the cap it clamps, because asking for more than there is is not an error.

One meaning for the parameter across the route, rather than a Feed-only one: it lowers the Name Query's cap the same way. The category listing is the exception and says so out loud — it answers everything it finds, so it has no cap to lower, and a `limit` there is a 400 rather than a parameter silently dropped. Giving the category listing a cap of its own would truncate answers clients get in full today; that is a separate change with its own migration path.

## Consequences

Every Shared Recipe is Feed-eligible. There is no editorial layer, no minimum quality bar and no way to exclude a Recipe from the Feed while leaving it shared, so a corpus that fills with low-effort Recipes surfaces them first by construction. The catalog ADR-0002 names as its destination is where a curation signal would live.

Recency-ordered means a Feed that changes only when someone shares. A client polling it during a quiet week gets the same twenty Recipes each time; the rotation the app shows on top of that stays the client's business.

The ordering is backed by `recipes_feed_idx` on `(created_at DESC NULLS LAST, id DESC)`, whose shape matches the query's `ORDER BY` exactly so the `LIMIT` stops at the last row asked for rather than sorting the corpus.
