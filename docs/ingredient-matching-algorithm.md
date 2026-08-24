# Ingredient matching algorithm

This document specifies how a set of Pantry Items is matched against Recipes and how the resulting Matches are ranked. Per ADR-0001 the algorithm is implemented twice — in Go against the Postgres corpus of Shared Recipes, and in Swift against the SwiftData store of Private and Saved Recipes — and the client merges the two ranked lists. Neither implementation is authoritative. **This document is.** If Go and Swift disagree, at least one of them is wrong, and this document says which.

The fixture table in [Fixtures](#fixtures) is the mechanism that keeps them in step. Both implementations consume it as test data. It is not illustrative; it is the contract.

Vocabulary is from `CONTEXT.md` and is used exactly: Pantry Item, Ingredient, Ingredient Term, Recipe, Match, Coverage, Missing Ingredient, Staple.

---

## 1. Normalization

Normalization turns one raw string — a Pantry Item as the user typed it, or an Ingredient's `name` as the Author wrote it — into an **Ingredient Term**, which for the purposes of this algorithm is a *set of tokens*.

Apply these steps in this order. The order matters and is part of the specification.

1. **Lowercase.** Simple, locale-independent Unicode lowercasing. Go: `strings.ToLower`. Swift: `lowercased()` — note Swift's default is locale-independent, which is what is wanted; do not pass a locale.
2. **Split on non-alphanumeric boundaries.** Scan the string and cut at every character that is not a Unicode letter (categories `L*`) or a Unicode decimal digit (category `Nd`). Every run of letters and digits between boundaries is one candidate token. Consecutive boundary characters produce no empty tokens. This means hyphens, periods, commas, slashes, parentheses, and all whitespace are separators, and it means Unicode fractions such as `½` (category `No`, not `Nd`) are separators rather than digits.
3. **Strip digits from each token.** Remove every `Nd` character from each candidate token. `2` becomes empty; `500g` becomes `g`.
4. **Drop empty tokens.**
5. **Drop stopwords.** Remove any token present in `DescriptorStopwords` or in `MeasurementStopwords` (section 2).
6. **Deduplicate.** The result is a set, not a list. Order is never significant anywhere in this algorithm.

The result may be the empty set. That is a legal outcome and section 5 says what to do with it.

Step 2 is stated as "split, then strip" rather than "strip, then split" deliberately. Stripping punctuation first would fuse `fl.oz` into the single token `floz` and `salt/pepper` into `saltpepper`; splitting first yields `fl`, `oz` and `salt`, `pepper`, which is what the stopword lists and the matcher expect.

Colour and variety words are **not** stopwords. `yellow`, `red`, `green`, `sweet`, `baby` and the like stay in the token set. They cost nothing — matching is per token (section 4), so `2 Yellow Onions` still matches `onion` through the `onions`/`onion` pair — and removing them would erase real distinctions such as `sweet potato` and `green onion`.

### Examples

| Raw string | Ingredient Term (token set) |
| --- | --- |
| `Tomatoes` | `{tomatoes}` |
| `  ToMaTo  ` | `{tomato}` |
| `2 Yellow Onions` | `{yellow, onions}` |
| `Boneless, skinless chicken breasts` | `{chicken, breasts}` |
| `1 1/2 cups all-purpose flour` | `{all, purpose, flour}` |
| `Salt & pepper, to taste` | `{salt, pepper}` |
| `freshly chopped` | `{}` (empty) |
| `` / `   ` | `{}` (empty) |

---

## 2. Stopword lists

Two lists, kept separate and named separately in both implementations, because they have different provenance and will change for different reasons.

### 2.1 `MeasurementStopwords`

**Provenance: transcribed from the `aliases` table of the `MeasurementUnit` enum in `Bitely-iOS/Bitely/Models/Ingredient.swift`.** That enum is currently commented out — it was written for the shopping list feature and kept — but it is the existing, already-reviewed measurement vocabulary in this product, and issue 3's review comment records why it is the source: deriving a second measurement list independently would guarantee the Go and Swift implementations disagree on exactly the inputs the fixture table exists to protect.

Multi-word aliases (`fl oz`, `fluid ounce`, `fl. oz`, `to taste`) are split into their component tokens here, because normalization has already tokenized by the time stopwords are applied. The `MeasurementUnit.none` alias `[""]` contributes nothing.

```
// Volume
tsp, t, teaspoon, teaspoons
tbsp, tablespoon, tablespoons, tbs
fl, floz, fluid, ounce, ounces
cup, cups, c
pint, pints, pt
quart, quarts, qt
gallon, gallons, gal
ml, milliliter, milliliters, millilitre, millilitres
l, liter, liters, litre, litres

// Mass
g, gram, grams
kg, kilogram, kilograms
oz
lb, lbs, pound, pounds

// Count
piece, pieces, whole
clove, cloves
can, cans, tin, tins
slice, slices

// Special
pinch, pinches
dash, dashes
handful, handfuls
to, taste
```

Notes on transcription:

- The tablespoon alias `T` lowercases to `t`, colliding with the teaspoon alias `t`. Both are stopwords, so the collision is harmless. Do not try to preserve the case distinction; normalization lowercases first.
- `ounce`/`ounces` appear under both `fluidOunce` and `ounce`. The list is a set; one entry each.
- `oz` covers the mass unit; `floz` covers the collapsed alias `fl. oz` written without the boundary. Both are needed because normalization splits `fl. oz` into `fl` + `oz` but a user typing `floz` produces one token.
- Single-letter entries (`t`, `c`, `l`, `g`) are aggressive but safe: no Ingredient Term of value is a single letter, and dropping them prevents `1c` from surviving as the token `c`.

When the `MeasurementUnit` enum changes, this list changes, and both implementations change with it. See section 8.

### 2.2 `DescriptorStopwords`

Preparation and size words that describe how a food arrives rather than what it is, plus the handful of function words that appear inside ingredient lines. Kept in one list rather than split further because they are dropped for the same reason and are tuned together.

```
// Preparation
chopped, diced, minced, sliced, shredded, grated, crushed, ground,
mashed, cubed, julienned, halved, quartered, trimmed, peeled, seeded,
pitted, stemmed, rinsed, drained, softened, melted, beaten, packed,
sifted, toasted, roasted, cooked, uncooked, raw, boneless, skinless

// Condition and quality
fresh, freshly, frozen, dried, ripe, unsalted, unsweetened, plain,
low, reduced, room, temperature, warm, cold, hot, boiling

// Size and amount
large, small, medium, extra, thin, thick, finely, coarsely, thinly,
thickly, lightly, heaping, scant

// Line noise
optional, divided, plus, more, needed, garnish, and, or, of, the, a,
an, for, with, into, in, on, about, approximately
```

`boneless` and `skinless` are here because the example rows in section 1 and section 6.2 always assumed they were, and because they are the two commonest descriptors on a packaged cut of meat. Leaving them out changed no MATCH verdict — nobody submits `boneless` as a Pantry Item, and neither token scores against a food — but it did change Coverage: `Chicken breasts` and `Boneless, skinless chicken breasts` on one Recipe are one Ingredient Term with them dropped and two without, and the denominator decides the rank.

Explicitly **not** stopwords, and each for a reason:

- `salted`, `sweetened`, `fat`, `cream`, `stock`, `broth` — these are foods or distinguish foods.
- `sweet`, `baby`, `wild`, `whole` grain qualifiers — except `whole`, which is already a `MeasurementUnit.piece` alias and is dropped by that list. This is a known small loss: `whole wheat flour` normalizes to `{wheat, flour}`. Acceptable, since `flour` still matches.
- Colour words — see section 1.

---

## 3. Similarity

Similarity between two **tokens** is a faithful port of the definition Postgres `pg_trgm` uses, so that the Go implementation, the Swift implementation, and any SQL written against the corpus agree.

Given a token `s`:

1. Prepend exactly **two** space characters and append exactly **one** space character. `onion` becomes `"  onion "`.
2. Take every contiguous substring of length 3 at every offset from 0 to `len - 3` inclusive, where `len` is the length of the padded string. A padded token of length `n + 3` yields `n + 1` extracted trigrams for an original token of length `n`.
3. Collect them into a **set**. Duplicate trigrams collapse — this is a set, not a multiset, and the count of distinct trigrams is what the arithmetic uses. `banana` extracts `bana`… specifically `ana` twice; it contributes one element.

Then, for tokens `a` and `b` with trigram sets `A` and `B`:

```
similarity(a, b) = |A ∩ B| / |A ∪ B|
```

Both sides are integers; the quotient is a real number in `[0, 1]`. Identical tokens score exactly `1.0`. Tokens sharing no trigram score exactly `0.0`.

Worked example — `onion` vs `onions`, writing `_` for a space:

```
A = trigrams("  onion ") = { __o, _on, oni, nio, ion, on_ }        |A| = 6
B = trigrams("  onions ") = { __o, _on, oni, nio, ion, ons, ns_ }  |B| = 7
A ∩ B = { __o, _on, oni, nio, ion }                                 5
A ∪ B = { __o, _on, oni, nio, ion, on_, ons, ns_ }                  8
similarity = 5 / 8 = 0.625
```

Two notes on fidelity to `pg_trgm`:

- Postgres applies this to whole strings by first lowercasing, replacing non-alphanumerics with spaces, and padding each *word* the same way, then unioning the per-word trigram sets. Because this algorithm always compares single tokens that normalization has already produced, the two definitions coincide on our inputs. They do **not** coincide on multi-word strings, which is why section 4 forbids whole-string comparison.
- `pg_trgm` also de-duplicates. `similarity('banana', 'banana')` is `1.0`, not something less. Any implementation using multisets will disagree with Postgres and with this document.

Floating-point: compute `|A ∩ B|` and `|A ∪ B|` as integers and compare against the threshold as `intersection * 10 >= union * 3` if exact integer comparison is preferred. `0.3` is not exactly representable in binary floating point, and `egg`/`eggplant` lands exactly on the threshold (section 6), so this is not hypothetical.

---

## 4. Comparison is per token

A Pantry Item **matches** an Ingredient Term when **any** token from the Pantry Item's token set scores at or above the threshold against **any** token from the Ingredient Term's token set.

```
matches(P, T)  ⟺  ∃ p ∈ P, ∃ t ∈ T : similarity(p, t) >= MatchThreshold
```

This is the central correction recorded in ADR-0002. Whole-string comparison fails the motivating case. `chicken breast` against `boneless skinless chicken breasts` compared as strings drags in every trigram of `boneless` and `skinless`; those trigrams are all in the union and none are in the intersection, so the similarity falls far below any threshold that is not also low enough to match unrelated foods. Per-token comparison finds the `chicken`/`chicken` pair at `1.0` and stops.

The cost is stated plainly in ADR-0002 and asserted in the fixtures: `chicken breast` also matches `chicken stock`, because `chicken`/`chicken` is `1.0` there too. That is a false positive we accept. It is the trade we made for shipping without the Ingredient Catalog.

---

## 5. Coverage and ranking

### Constants

```
MatchThreshold = 0.3
```

One named constant, in both implementations, with this name. The match is **binary**: a token pair at `0.99` and a token pair at `0.31` both produce exactly one matched Ingredient Term. There is no fractional credit, because Coverage must remain an integer ratio so the interface can say "you have 4 of 6 ingredients" — a sentence fractional credit makes unsayable.

### Building a Recipe's Ingredient Term set

1. Normalize each Ingredient's `name` into an Ingredient Term.
2. **Discard Ingredients whose Ingredient Term is empty.** An Ingredient named `to taste` or `chopped` contributes nothing, is not counted in the denominator, and can never be a Missing Ingredient.
3. **Deduplicate by token set.** If two Ingredients on one Recipe normalize to the same Ingredient Term, they count once. `2 large onions` and `1 small onion` on the same Recipe are one Ingredient Term.

The `measurement` field is never read. Quantity is ignored entirely: a Pantry Item asserts only that the user has *some* of a food, and comparing "I have flour" against "needs 6 cups" would need a unit system this product does not have.

### Building the Pantry Item set

Normalize each submitted string, discard the ones that normalize to empty, and deduplicate by token set.

### Coverage

```
Coverage = |matched Ingredient Terms| / |Recipe's Ingredient Terms|
```

An Ingredient Term is **matched** if at least one Pantry Item matches it (section 4), and is a **Missing Ingredient** otherwise. A Recipe with zero Ingredient Terms after the discard rule has Coverage `0` and is not returned as a Match.

A Recipe the Pantry covers none of is not returned either. Narrowing is looser than scoring (section 8), so a candidate can arrive with nothing matched, and a Match with Coverage `0` offers the user a Recipe they hold no Ingredient for. Both implementations drop it.

**Staples are not excluded.** The term is defined in `CONTEXT.md`, but identifying Staples reliably needs the deferred Ingredient Catalog, and ADR-0002 records that a hardcoded list was rejected as a second source of truth in two languages. Until the catalog exists, salt counts against Coverage like anything else.

**There is no cap on Missing Ingredients.** A cap combined with unrecognized Staples would hide good Matches behind "missing salt, missing pepper". Coverage ordering alone buries weak Matches.

### Ranking

Sort Matches by, in order:

1. **Coverage, descending.** Compare as integers to avoid float divergence between languages: for Matches `x` and `y`, `x` sorts first when `x.matched * y.total > y.matched * x.total`.
2. **Missing Ingredient count, ascending.** Fewer missing wins.
3. **Recipe name, ascending, by UTF-8 byte order.** Go's `<` on `string` is exactly this. Swift's `<` on `String` is *not* — it compares by Unicode canonical equivalence — so Swift must compare `Array(a.utf8)` against `Array(b.utf8)` with `lexicographicallyPrecedes`. Getting this wrong produces two differently-ordered lists for names differing only in accent composition.
4. **Recipe id, ascending, by the same UTF-8 byte rule.** A final total order, so the merged list is deterministic.

Steps 1 and 2 encode user story 7: a twelve-Ingredient Recipe missing three (`9/12 = 0.75`) outranks a four-Ingredient Recipe missing three (`1/4 = 0.25`), because being close matters more than being small.

---

## 6. Fixtures

Both implementations consume this table as test data. Every row is an assertion.

### 6.1 Token-pair similarity

Arithmetic computed by hand from section 3. `|A|`, `|B|` are distinct trigram counts; `I` and `U` are intersection and union sizes. Verdict is `sim >= 0.3`.

| Token A | Token B | \|A\| | \|B\| | I | U | sim | Verdict | Protects |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `tomato` | `tomato` | 7 | 7 | 7 | 7 | 1.0000 | MATCH | Identity is exactly 1.0 |
| `tomato` | `tomatoes` | 7 | 9 | 6 | 10 | 0.6000 | MATCH | The plural case from user story 8 |
| `onion` | `onions` | 6 | 7 | 5 | 8 | 0.6250 | MATCH | Plural, worked in full in section 3 |
| `breast` | `breasts` | 7 | 8 | 6 | 9 | 0.6667 | MATCH | Plural inside a multi-token term |
| `potato` | `potatoes` | 7 | 9 | 6 | 10 | 0.6000 | MATCH | `-o`/`-oes` plural |
| `egg` | `eggs` | 4 | 5 | 3 | 6 | 0.5000 | MATCH | Short tokens still clear the bar |
| `oat` | `oats` | 4 | 5 | 3 | 6 | 0.5000 | MATCH | Three-letter token |
| `mushroom` | `mushrooms` | 9 | 10 | 8 | 11 | 0.7273 | MATCH | Long token, plural |
| `yogurt` | `yoghurt` | 7 | 8 | 5 | 10 | 0.5000 | MATCH | Spelling variant, not a plural |
| `egg` | `eggplant` | 4 | 9 | 3 | 10 | **0.3000** | MATCH | **Boundary.** Exactly at threshold; the comparison is `>=`, not `>`. Any implementation that computes `3/10` in floating point and tests `> 0.3` will disagree here. Use the integer form. |
| `rice` | `ricotta` | 5 | 8 | 3 | 10 | **0.3000** | MATCH | Second boundary case, same reason |
| `chicken` | `chickpea` | 8 | 9 | 5 | 12 | **0.4167** | **MATCH** | **See section 7.1 — this was expected to be a non-match and is not.** |
| `chicken` | `chickpeas` | 8 | 10 | 5 | 13 | 0.3846 | MATCH | Same family as above; asserted so the behaviour is recorded |
| `chicken` | `kitchen` | 8 | 8 | 1 | 15 | 0.0667 | no match | Anagram-ish token that must not match |
| `chicken` | `chili` | 8 | 6 | 3 | 11 | 0.2727 | no match | Same first letters, below threshold |
| `tomato` | `potato` | 7 | 7 | 2 | 12 | 0.1667 | no match | Rhyming foods stay distinct |
| `basil` | `basmati` | 6 | 8 | 3 | 11 | 0.2727 | no match | Near-miss just under the bar; moving the threshold to 0.25 breaks this row |
| `lemon` | `lime` | 6 | 5 | 1 | 10 | 0.1000 | no match | Related foods, unrelated spellings |
| `beef` | `broth` | 5 | 6 | 1 | 10 | 0.1000 | no match | Words from the same recipe line |
| `kale` | `cake` | 5 | 5 | 0 | 10 | 0.0000 | no match | Same letters, no shared trigram |
| `chicken` | `stock` | 8 | 6 | 0 | 14 | 0.0000 | no match | The *other* token of the false-positive pair scores zero |
| `yellow` | `onion` | 7 | 6 | 0 | 13 | 0.0000 | no match | Descriptor token contributes nothing on its own |
| `butter` | `buttermilk` | 7 | 11 | 6 | 12 | 0.5000 | MATCH | Prefix containment matches |
| `milk` | `buttermilk` | 5 | 11 | 3 | 13 | 0.2308 | no match | Suffix containment does **not**, because of the two leading pad spaces. This asymmetry is real, is inherited from `pg_trgm`, and must not be "fixed". |
| `bread` | `breadcrumbs` | 6 | 12 | 5 | 13 | 0.3846 | MATCH | Prefix containment again |
| `corn` | `cornstarch` | 5 | 11 | 4 | 12 | 0.3333 | MATCH | Prefix containment, near the bar |
| `apple` | `pineapple` | 6 | 10 | 4 | 12 | 0.3333 | MATCH | Compound food, accepted |
| `pepper` | `pepperoni` | 7 | 10 | 6 | 11 | 0.5455 | MATCH | Known false positive, prefix family |
| `pea` | `peanut` | 4 | 7 | 3 | 8 | 0.3750 | MATCH | Known false positive, short token |
| `celery` | `celeriac` | 7 | 9 | 5 | 11 | 0.4545 | MATCH | Known false positive |
| `parsley` | `parsnip` | 8 | 8 | 4 | 12 | 0.3333 | MATCH | Known false positive, just over the bar |
| `onion` | `union` | 6 | 6 | 3 | 9 | 0.3333 | MATCH | Not a food pair, but pins the arithmetic |
| `beef` | `beets` | 5 | 6 | 3 | 8 | 0.3750 | MATCH | Known false positive between real foods |

### 6.2 Pantry Item against Ingredient Term

Left column is the raw Pantry Item string as typed. Middle is the raw Ingredient `name` as an Author wrote it. Verdict is section 4 applied to the normalized token sets.

| Pantry Item (raw) | Ingredient name (raw) | Normalized pair | Verdict | Protects |
| --- | --- | --- | --- | --- |
| `tomato` | `tomato` | `{tomato}` / `{tomato}` | MATCH | Exact match |
| `Tomato` | `tomato` | `{tomato}` / `{tomato}` | MATCH | Casing (user story 10) |
| `  ToMaTo  ` | `Tomato` | `{tomato}` / `{tomato}` | MATCH | Leading/trailing whitespace and mixed case |
| `tomato` | `Tomatoes` | `{tomato}` / `{tomatoes}` | MATCH | Plural, sim 0.6 (user story 8) |
| `Tomatoes` | `1 tomato, diced` | `{tomatoes}` / `{tomato}` | MATCH | Plural in the other direction, quantity and descriptor stripped from the Ingredient |
| `chicken breast` | `boneless skinless chicken breasts` | `{chicken, breast}` / `{chicken, breasts}` | MATCH | **The motivating case** (user story 9). Matches on `chicken`/`chicken` = 1.0. Would fail whole-string comparison. |
| `chicken breast` | `Boneless, Skinless Chicken Breasts` | `{chicken, breast}` / `{chicken, breasts}` | MATCH | Same, with punctuation and casing |
| `2 Yellow Onions` | `onion` | `{yellow, onions}` / `{onion}` | MATCH | Quantity-prefixed input (user story 11). Matches on `onions`/`onion` = 0.625; `yellow`/`onion` = 0.0 and contributes nothing. |
| `1 1/2 cups all-purpose flour` | `Flour` | `{all, purpose, flour}` / `{flour}` | MATCH | Unicode-free fraction, measurement stopword, hyphen split |
| `½ cup ⅓ milk` | `Milk` | `{milk}` / `{milk}` | MATCH | Unicode fractions are boundary characters, not digits |
| `500g Beef` | `ground beef` | `{beef}` / `{beef}` | MATCH | Digit-fused unit (`500g` → `g` → dropped as a measurement stopword); `ground` dropped as a descriptor |
| `chicken breast` | `chicken stock` | `{chicken, breast}` / `{chicken, stock}` | **MATCH** | **The known false positive, asserted deliberately.** ADR-0002 records this as the price of shipping without the Ingredient Catalog. Do not "fix" this row by raising `MatchThreshold`; raising it to 0.4 breaks `corn`/`cornstarch`, `apple`/`pineapple` and `parsley`/`parsnip` in the other direction and does nothing here, because the offending pair scores 1.0. |
| `heavy cream` | `sour cream` | `{heavy, cream}` / `{sour, cream}` | MATCH | Same false-positive family: the shared head token decides it |
| `olive oil` | `vegetable oil` | `{olive, oil}` / `{vegetable, oil}` | MATCH | Same again, and the case a Staples list would eventually make moot |
| `chicken` | `chickpeas` | `{chicken}` / `{chickpeas}` | **MATCH** | **Expected to be a non-match; is not.** See section 7.1. |
| `chicken` | `chicken thighs` | `{chicken}` / `{chicken, thighs}` | MATCH | One token of a Pantry Item against a multi-token Ingredient Term |
| `tomato` | `potato` | `{tomato}` / `{potato}` | no match | Near-miss that must not match |
| `basil` | `basmati rice` | `{basil}` / `{basmati, rice}` | no match | Both token pairs below the bar: 0.2727 and 0.0 |
| `beef` | `chicken broth` | `{beef}` / `{chicken, broth}` | no match | No token pair clears the bar |
| `` (empty) | `tomato` | `{}` / `{tomato}` | no match | Empty input matches nothing |
| `   ` (whitespace) | `tomato` | `{}` / `{tomato}` | no match | Whitespace-only input matches nothing |
| `,.-/()` | `tomato` | `{}` / `{tomato}` | no match | Punctuation-only input normalizes to empty |
| `2 1/2` | `tomato` | `{}` / `{tomato}` | no match | Digits-only input normalizes to empty |
| `freshly chopped` | `fresh chopped tomato` | `{}` / `{tomato}` | no match | Pantry Item that is entirely stopwords matches nothing — including an Ingredient that shares those stopwords |
| `to taste` | `salt` | `{}` / `{salt}` | no match | Measurement stopwords alone normalize to empty |
| `tomato` | `to taste` | `{tomato}` / `{}` | no match | An Ingredient that normalizes to empty is not merely unmatchable — it is discarded from the Recipe's Ingredient Term set entirely (section 5) |
| `Salt & Pepper` | `black pepper` | `{salt, pepper}` / `{black, pepper}` | MATCH | Ampersand as a boundary; second token carries the match |

### 6.3 Recipe Coverage and ranking

**Scenario A — user story 7.** Pantry Items: `chicken`, `onion`, `garlic`, `carrot`, `celery`, `salt`, `pepper`, `bay leaves`, `noodles`.

| Recipe | Ingredient Terms | Matched | Missing | Coverage |
| --- | --- | --- | --- | --- |
| `Chicken Noodle Soup` | chicken, onion, carrot, celery, garlic, salt, pepper, bay, noodles, thyme, parsley, butter (12) | 9 | thyme, parsley, butter (3) | 9/12 = 0.75 |
| `Garlic Bread` | garlic, butter, bread, parsley (4) | 1 | butter, bread, parsley (3) | 1/4 = 0.25 |

Assertion: `Chicken Noodle Soup` ranks **above** `Garlic Bread`. Both are missing exactly three Ingredient Terms; Coverage is the primary sort and decides it. This is the row that fails if anyone ranks by absolute matched count or by absolute missing count.

Also assert on this scenario: `bay leaves` normalizes to `{bay, leaves}` and matches the `bay` term; `Chicken Noodle Soup`'s Missing Ingredients are reported as exactly `thyme`, `parsley`, `butter`, in the Recipe's own Ingredient order.

**Scenario B — full Coverage first, then the tie-breaks.** Pantry Items: `eggs`, `butter`, `flour`, `sugar`, `milk`.

| Recipe | Ingredient Terms | Matched | Missing | Coverage | Expected rank |
| --- | --- | --- | --- | --- | --- |
| `Butter Cookies` | butter, sugar, flour, egg (4) | 4 | — (0) | 4/4 = 1.00 | 1 |
| `Crepes` | eggs, flour, milk, butter (4) | 4 | — (0) | 4/4 = 1.00 | 2 |
| `Pound Cake` | butter, sugar, eggs, flour, vanilla, salt (6) | 4 | `vanilla`, `salt` (2) | 4/6 ≈ 0.667 | 3 |
| `Pancakes` | flour, sugar, eggs, milk, butter, baking powder, salt, oil (8) | 5 | `baking powder`, `salt`, `oil` (3) | 5/8 = 0.625 | 4 |
| `Waffles` | flour, eggs, milk, butter, sugar, baking soda, salt, cinnamon (8) | 5 | `baking soda`, `salt`, `cinnamon` (3) | 5/8 = 0.625 | 5 |

Assertions on Scenario B:

- `Crepes` and `Butter Cookies` both have Coverage 1.00 and 0 Missing Ingredients, so the tie falls to Recipe name. `Butter Cookies` precedes `Crepes` in UTF-8 byte order, so it ranks first. A run that puts `Crepes` first has either skipped the name tie-break or sorted unstably.
- Both fully covered Recipes rank above every partially covered one, which is user story 6.
- `Butter Cookies` matching its `egg` Ingredient Term against the Pantry Item `eggs` exercises the plural path inside a Coverage computation rather than in isolation.
- `Pancakes` and `Waffles` tie on Coverage `5/8` and on Missing Ingredient count (3), so the name tie-break orders `Pancakes` before `Waffles`.
- No Recipe is excluded for having three Missing Ingredients. There is no cap.

**Scenario C — the equal-Coverage, unequal-denominator tie-break.** Pantry Items: `rice`, `onion`, `egg`.

| Recipe | Ingredient Terms | Matched | Missing | Coverage | Expected rank |
| --- | --- | --- | --- | --- | --- |
| `Rice Pilaf` | rice, onion, broth, butter (4) | 2 | 2 | 2/4 = 0.5 | 1 |
| `Fried Rice` | rice, onion, egg, soy sauce, oil, peas (6) | 3 | 3 | 3/6 = 0.5 | 2 |
| `Onion Soup` | onion, broth, bread, cheese, butter, thyme (6) | 1 | 5 | 1/6 ≈ 0.167 | 3 |

Assertions on Scenario C:

- `Rice Pilaf` (2 of 4) and `Fried Rice` (3 of 6) have identical Coverage. The integer comparison confirms it exactly: `2 * 6 == 3 * 4`. Implementations that compare `0.5` computed as a float may or may not agree; this is precisely why tie-break 1 is specified as cross-multiplication.
- The tie resolves on Missing Ingredient count: 2 missing beats 3 missing, so `Rice Pilaf` ranks first. This is the only scenario in the table that exercises tie-break 2 in isolation, with tie-break 1 exactly equal and tie-break 3 not reached.

**Scenario D — degenerate inputs.**

| Input | Expected |
| --- | --- |
| Every submitted string is blank (`""`, `"   "`), or the list itself is empty (`[]`) | `400`. A blank names no food, so such a request carries no Pantry Item at all and is malformed rather than a pantry that matches nothing. See ADR-0003. |
| Pantry Items all normalize to empty without being blank (`"to taste"`, `"freshly chopped"`) | No Matches: `200` with an empty list. A blank sitting alongside such a Pantry Item is discarded and does not make the request a `400`. |
| Pantry Items contain duplicates and blanks (`"Onion"`, `"onion"`, `""`, `" onions "`) | Deduplicated to `{onion}` and `{onions}` — two distinct Ingredient Terms, both of which match an `onion` Ingredient Term, contributing one matched Ingredient Term, not two. Coverage counts Ingredient Terms on the Recipe, never Pantry Items. |
| Recipe with every Ingredient normalizing to empty | Zero Ingredient Terms, Coverage 0, not returned as a Match. Never a division by zero. |

---

## 7. Known divergences from the stated expectations

### 7.1 `chicken` / `chickpea` matches, and was not supposed to

The design notes list `chicken`/`chickpea` as a near-miss that must **not** match. The arithmetic says otherwise:

```
A = trigrams("  chicken ")  = { __c, _ch, chi, hic, ick, cke, ken, en_ }     |A| = 8
B = trigrams("  chickpea ") = { __c, _ch, chi, hic, ick, ckp, kpe, pea, ea_ } |B| = 9
A ∩ B = { __c, _ch, chi, hic, ick }                                            5
A ∪ B = 12
similarity = 5 / 12 = 0.4167
```

`0.4167 >= 0.3`, so it matches. The five-character shared prefix `chick`, plus the two-space pad that makes prefixes count double, is enough on its own. `chicken`/`chickpeas` is `0.3846` and matches for the same reason.

The fixture table records the **real** verdict, MATCH, not the intended one. Fudging the expectation would put a failing assertion in both test suites; fudging the threshold to exclude it would need `MatchThreshold > 0.4167`, which would also drop `corn`/`cornstarch` (0.3333), `apple`/`pineapple` (0.3333), `parsley`/`parsnip` (0.3333) and `bread`/`breadcrumbs` (0.3846) — every one of which is either desirable or in the same accepted-imprecision bucket.

This is not a new class of problem. It is the same trade ADR-0002 already made and named: precision is what we gave up to ship without the Ingredient Catalog. `chicken`/`chickpea` belongs alongside `chicken breast`/`chicken stock` as a false positive the catalog is expected to buy back. It is asserted as a MATCH so that the behaviour is recorded rather than rediscovered.

### 7.2 Prefix containment is asymmetric

`butter`/`buttermilk` is `0.5` and matches; `milk`/`buttermilk` is `0.2308` and does not. The two leading pad spaces give the head of a token two trigrams that exist nowhere else, so a shared prefix is worth much more than a shared suffix. This is inherited from `pg_trgm` and is correct behaviour for this specification. Both rows are in the fixture table so that nobody "corrects" the asymmetry into existence or out of it.

---

## 8. Where each piece runs

### Postgres narrows; it never scores

Postgres's only job is to reduce the corpus to a candidate set. It does not compute Coverage, does not order results, and its similarity verdicts are not trusted. The pure Go matching package re-scores **every** candidate from scratch using this document's rules. This keeps scoring in one testable place with no database dependency, and it means a bug in the SQL costs recall, never correctness.

Candidate narrowing is deliberately **looser** than final scoring: a Recipe is a candidate if it has at least one Ingredient whose `name_norm` is trigram-similar to at least one submitted term. Candidates are capped at a generous ceiling before scoring, so a pantry of common foods cannot pull the whole corpus into memory.

`ingredients.name_norm` is a generated column, `lower(btrim(name, E' \t\r\n'))` only — it is not the normalization in section 1, and must not be mistaken for it. Postgres derives it from `name`, so no writer can leave it NULL or set it inconsistently. It exists as a narrowing surface, nothing more.

### Migration 013's index is the wrong index

Migration 013 created:

```sql
CREATE INDEX ingredients_name_norm_idx ON ingredients(name_norm);
```

That is a btree index. Btree serves equality and prefix range scans. It cannot serve trigram similarity — the `%` operator and `similarity()` will sequential-scan straight past it. It is not a substitute and it is not a starting point.

Narrowing requires:

```sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX ingredients_name_norm_trgm_idx
  ON ingredients USING gin (name_norm gin_trgm_ops);
```

The btree index can stay; it costs a little write throughput and serves nothing this feature needs.

### The Swift side

The Swift implementation reads Private and Saved Recipes from SwiftData, applies this same document, and merges its ranked list with the API's. Because both lists were produced by this algorithm, the merge is a simple ordered merge on the same comparison keys — Coverage, then Missing Ingredient count, then name, then id — followed by deduplication of a Saved Recipe against its corpus original by `Recipe.remoteId`.

---

## 9. Changing the algorithm

The ranking function exists twice by design. The two drifting apart is the main long-term risk in this feature, and this document plus its fixture table is the only thing preventing it.

Any change to **normalization**, to either **stopword list**, or to **`MatchThreshold`** lands here first, and in both implementations together:

1. Change this document, including recomputing every affected fixture row by hand.
2. Change the Go implementation and the Swift implementation in the same review cycle. A change that ships to one side only makes the merged list incoherent, and it does so silently — no error, no crash, just a list where a Recipe from the corpus is ranked by different rules than the Recipe below it from the device.
3. Both test suites read the updated fixture table. A red test on either side blocks both.

The single change most likely to be proposed is raising `MatchThreshold` to reduce false positives. Section 7.1 shows why that does not work: the false positives that hurt most (`chicken breast`/`chicken stock`, `heavy cream`/`sour cream`) score `1.0` on a shared token and are immune to any threshold below `1.0`, while the threshold's real effect is to start dropping the plural and prefix matches the feature exists to catch. The fix for precision is the Ingredient Catalog recorded in ADR-0002, not a bigger number here.

Background: `CONTEXT.md`, ADR-0001, ADR-0002, issue 3.
