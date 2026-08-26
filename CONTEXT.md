# Bitely

Bitely lets people write down recipes, keep them private on their device, and share the ones they want public. This context covers the API and the shared recipe corpus it owns.

## Language

### Recipes

**Recipe**:
A named set of instructions plus the ingredients it calls for.

**Shared Recipe**:
A Recipe published to the Bitely corpus, readable by anyone. Publishing to the corpus *is* the act of sharing; there is no private recipe on the server.
_Avoid_: Public recipe, published recipe

**Private Recipe**:
A Recipe that exists only in the local store on the author's device and has never been shared.
_Avoid_: Draft, unpublished recipe

**Saved Recipe**:
Another person's Shared Recipe that a user has kept a copy of on their device. The same Recipe exists in both the corpus and the user's local store, and the local copy carries the corpus Recipe's identity so the two can be recognized as one.
_Avoid_: Bookmarked recipe, favorite

**Author**:
The user who created a Recipe. The only user permitted to change or delete it once shared.
_Avoid_: Owner, creator

### Ingredients

**Ingredient**:
A single line on a Recipe: what it calls for and how much. Belongs to exactly one Recipe and is never shared between Recipes.
_Avoid_: Recipe ingredient, ingredient line

**Ingredient Term**:
The normalized food name an Ingredient refers to, stripped of quantity and formatting. What matching compares. Several Ingredients across different Recipes can share one Ingredient Term.
_Avoid_: Ingredient name, normalized name

**Pantry Item**:
A food a user says they currently have, without quantity: it asserts only that the user has some. Compared against Ingredient Terms to find Recipes they could cook.
_Avoid_: Available ingredient, user ingredient, on-hand item

**Staple**:
An Ingredient Term so commonly stocked that its absence should not count against a Recipe. Salt, water, cooking oil.
_Avoid_: Basic, common ingredient

### Matching

**Match**:
A Recipe offered in response to a set of Pantry Items, together with which of its Ingredient Terms the user has and which they lack. A Match need not be complete.
_Avoid_: Suggestion, hit, result

**Coverage**:
The fraction of a Recipe's Ingredient Terms the user holds as Pantry Items, counting Staples only while they remain unidentified. The primary ordering for Matches.
_Avoid_: Score, match percentage

**Missing Ingredient**:
An Ingredient Term a Recipe calls for that the user has no matching Pantry Item for.
_Avoid_: Gap, shortfall

**Feed**:
The bounded selection of Shared Recipes offered to a user who has named nothing to narrow by — no category, no Name Query, no Pantry Items. Ordered by recency of sharing, because that is the only signal the corpus measures.
_Avoid_: Discover, home feed, today's picks, trending

**Name Query**:
What a user types to reach a Shared Recipe they already have in mind. Compared against Recipe names only, never Ingredients, and compared loosely enough to survive a misspelling. Distinct from a Match, which the user does not name.
_Avoid_: Search term, keyword, title search
