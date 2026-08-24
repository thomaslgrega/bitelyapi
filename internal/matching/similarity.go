package matching

// MatchThreshold is the similarity two tokens must reach to be treated as the
// same food. Section 5 of docs/ingredient-matching-algorithm.md names this
// constant and fixes its value at 0.3; section 9 explains why raising it does
// not buy the precision it looks like it would.
//
// Nothing compares against this float. It is here because both implementations
// are required to name the threshold, and because a client may want to render
// it. The comparison itself is clearsThreshold, which stays in integers: 0.3
// has no exact binary representation, and `egg`/`eggplant` and `rice`/`ricotta`
// land exactly on it, so a float comparison is free to resolve those two rows
// differently in Go than in Swift.
const MatchThreshold = float64(thresholdNumerator) / float64(thresholdDenominator)

// The threshold as the integers the comparison actually uses. These are the
// single source of truth for its value; MatchThreshold is derived from them.
const (
	thresholdNumerator   = 3
	thresholdDenominator = 10
)

// A trigramSet is the distinct three-character substrings of one padded token.
// It is a set, not a multiset: `banana` yields `ana` twice and contributes it
// once, which is what makes similarity('banana', 'banana') exactly 1.0 in
// pg_trgm and here.
type trigramSet map[string]struct{}

// trigrams extracts a token's trigram set by the rule in section 3: pad with
// two leading spaces and one trailing space, then take every contiguous
// three-character window. The asymmetric padding is deliberate — it gives the
// head of a token two trigrams that exist nowhere else, which is why `butter`
// matches `buttermilk` and `milk` does not. Section 7.2 records that this must
// not be "corrected".
//
// Windows are counted in characters rather than bytes, so a token carrying a
// non-ASCII letter yields the same trigrams a Postgres UTF-8 database would.
func trigrams(token string) trigramSet {
	padded := []rune("  " + token + " ")

	set := make(trigramSet, len(padded))
	for i := 0; i+3 <= len(padded); i++ {
		set[string(padded[i:i+3])] = struct{}{}
	}
	return set
}

// overlap returns |A ∩ B| and |A ∪ B| for two trigram sets: the numerator and
// denominator of the ratio section 3 calls similarity. They come back as the
// two integers rather than as a quotient precisely so that nobody divides them
// — the comparison that matters is clearsThreshold, and it stays exact.
func (s trigramSet) overlap(other trigramSet) (intersection, union int) {
	smaller, larger := s, other
	if len(larger) < len(smaller) {
		smaller, larger = larger, smaller
	}

	for trigram := range smaller {
		if _, shared := larger[trigram]; shared {
			intersection++
		}
	}

	return intersection, len(s) + len(other) - intersection
}

// matches reports whether two trigram sets are similar enough to name the same
// food: section 4's per-pair predicate. It takes sets rather than tokens
// because scoring compares every pantry token against every Ingredient Term
// token, so a token's trigrams are extracted once and reused across every pair
// it takes part in.
func (s trigramSet) matches(other trigramSet) bool {
	return clearsThreshold(s.overlap(other))
}

// clearsThreshold applies MatchThreshold to a similarity expressed as its
// integer numerator and denominator. Cross-multiplying keeps the comparison in
// integers, and the test is `>=`, not `>`, so a pair landing exactly on the
// threshold matches.
func clearsThreshold(intersection, union int) bool {
	if union == 0 {
		return false
	}
	return intersection*thresholdDenominator >= union*thresholdNumerator
}
