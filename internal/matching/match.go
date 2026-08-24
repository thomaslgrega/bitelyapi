package matching

import "sort"

// A Candidate is a Recipe as the matcher sees it: an id, a name for the
// ranking tie-break, and its Ingredient names in whatever order the caller
// holds them.
// Nothing else about a Recipe affects scoring.
type Candidate struct {
	RecipeID        string
	Name            string
	IngredientNames []string
}

// A Match is a Recipe offered in response to a set of Pantry Items, together
// with which of its Ingredient Terms the user has and which they lack. Both
// lists name Ingredients as the Author wrote them, in the order the Candidate
// carried them. The Postgres corpus records no authored position, so that is
// storage order there rather than the order on the page.
type Match struct {
	RecipeID string
	Matched  []string
	Missing  []string
}

// Coverage is the fraction of the Recipe's Ingredient Terms the pantry holds.
// Ranking never uses this value — it compares the two counts as integers, in
// rankKey.precedes — but a client renders it.
func (m Match) Coverage() float64 {
	total := len(m.Matched) + len(m.Missing)
	if total == 0 {
		return 0
	}
	return float64(len(m.Matched)) / float64(total)
}

// Rank scores every candidate against the pantry and returns the Matches
// ordered best fit first: Coverage descending, then Missing Ingredient count
// ascending, then Recipe name, then Recipe id. Candidates the pantry covers
// none of are not Matches and are dropped.
func Rank(pantryItems []string, candidates []Candidate) []Match {
	held := newPantry(pantryItems)

	type ranked struct {
		match Match
		key   rankKey
	}

	results := make([]ranked, 0, len(candidates))
	for _, candidate := range candidates {
		match := held.score(candidate)
		if len(match.Matched) == 0 {
			continue
		}
		results = append(results, ranked{match: match, key: newRankKey(match, candidate.Name)})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].key.precedes(results[j].key)
	})

	matches := make([]Match, 0, len(results))
	for _, result := range results {
		matches = append(matches, result.match)
	}
	return matches
}

// A rankKey is everything ordering compares a Match on, and nothing else.
// Section 5 of the algorithm document specifies the comparison; it lives in
// its own type because it is the part of this package the Swift
// implementation has to mirror exactly, and because a comparison spread
// through a sort callback is one nobody can port line by line.
type rankKey struct {
	matched  int
	missing  int
	name     string
	recipeID string
}

func newRankKey(match Match, name string) rankKey {
	return rankKey{
		matched:  len(match.Matched),
		missing:  len(match.Missing),
		name:     name,
		recipeID: match.RecipeID,
	}
}

// total is the size of the Recipe's Ingredient Term set: the denominator of
// Coverage, and every Term is either matched or missing.
func (k rankKey) total() int {
	return k.matched + k.missing
}

// precedes reports whether this Match ranks ahead of another. The four keys of
// section 5, in order, and the last of them is the Recipe id, so the ordering
// is total: no two Matches are ever left genuinely tied, and the same corpus
// comes back in the same order every time.
func (k rankKey) precedes(other rankKey) bool {
	// Coverage descending, cross-multiplied rather than divided. Coverage is an
	// integer ratio, and two languages dividing the same two integers can land
	// on different last bits; multiplying keeps the comparison exact in both.
	coverage, otherCoverage := k.matched*other.total(), other.matched*k.total()
	if coverage != otherCoverage {
		return coverage > otherCoverage
	}

	if k.missing != other.missing {
		return k.missing < other.missing
	}

	// Recipe name, then Recipe id, both by UTF-8 byte order. Go's `<` on string
	// is exactly that. Swift's is not — it compares by Unicode canonical
	// equivalence, which calls two spellings of the same accent equal — so the
	// port compares `Array(a.utf8)` instead. Names differing only in accent
	// composition are where the two disagree.
	if k.name != other.name {
		return k.name < other.name
	}

	return k.recipeID < other.recipeID
}

// PantryTokens returns the distinct tokens a set of Pantry Items normalizes
// to, sorted. It is what narrowing searches the corpus on — narrowing needs
// the tokens, not the Terms, because it compares one word at a time.
func PantryTokens(items []string) []string {
	return distinctTokens(items)
}

// distinctTokens normalizes every raw string and returns the tokens they yield
// between them, deduplicated across the whole list and sorted.
func distinctTokens(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	tokens := make([]string, 0, len(items))

	for _, item := range items {
		for _, token := range Normalize(item).Tokens() {
			if _, duplicate := seen[token]; duplicate {
				continue
			}
			seen[token] = struct{}{}
			tokens = append(tokens, token)
		}
	}
	sort.Strings(tokens)

	return tokens
}

// a pantry is what the user holds, as the trigrams of its distinct tokens.
// Only the trigrams are kept, because scoring never looks at a pantry token as
// a string: it compares every one of them against every Ingredient Term token,
// so extracting each token's trigrams once per request rather than once per
// comparison is the whole reason this type exists.
type pantry struct {
	tokens []trigramSet
}

func newPantry(items []string) pantry {
	distinct := distinctTokens(items)

	held := pantry{tokens: make([]trigramSet, 0, len(distinct))}
	for _, token := range distinct {
		held.tokens = append(held.tokens, trigrams(token))
	}

	return held
}

// holds reports whether the pantry covers an Ingredient Term. Comparison is
// per token, never per whole string: a Term matches when any one Pantry Item
// token clears MatchThreshold against any one of its tokens, and the first
// such pair settles it. Section 4 explains why comparing the whole strings
// fails the case the feature exists for — the descriptors in `boneless
// skinless chicken breasts` flood the trigram union and bury the
// `chicken`/`chicken` pair that decides it.
//
// The verdict is binary. A pair at 0.99 and a pair at 0.31 both make the Term
// matched and nothing more, because Coverage has to stay an integer ratio.
func (p pantry) holds(term Term) bool {
	for _, token := range term.Tokens() {
		wanted := trigrams(token)
		for _, pantryToken := range p.tokens {
			if pantryToken.matches(wanted) {
				return true
			}
		}
	}
	return false
}

// score builds the candidate's Ingredient Term set — discarding Ingredients
// that normalize to empty, deduplicating by token set — and splits it into
// what the pantry covers and what it misses.
func (p pantry) score(candidate Candidate) Match {
	match := Match{RecipeID: candidate.RecipeID}

	seen := make(map[string]struct{}, len(candidate.IngredientNames))
	for _, name := range candidate.IngredientNames {
		term := Normalize(name)
		if term.IsEmpty() {
			continue
		}
		if _, duplicate := seen[term.key()]; duplicate {
			continue
		}
		seen[term.key()] = struct{}{}

		if p.holds(term) {
			match.Matched = append(match.Matched, name)
		} else {
			match.Missing = append(match.Missing, name)
		}
	}

	return match
}
