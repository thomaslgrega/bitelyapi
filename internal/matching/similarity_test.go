package matching

import (
	"math"
	"testing"
)

// Section 6.1 of docs/ingredient-matching-algorithm.md, transcribed row for
// row. Every column is asserted, not just the verdict: the distinct trigram
// counts and the intersection and union sizes are the arithmetic the Swift
// implementation must reproduce exactly, and a verdict can be right for the
// wrong reason.
//
// The verdict column goes through trigramSet.matches, which is the predicate
// scoring itself calls. A second spelling of the rule for the tests to hold
// could drift from the shipped one, and this table exists to catch drift.
func TestSection61Fixtures(t *testing.T) {
	tests := []struct {
		a, b         string
		countA       int
		countB       int
		intersection int
		union        int
		sim          float64
		match        bool
		protects     string
	}{
		{"tomato", "tomato", 7, 7, 7, 7, 1.0000, true, "identity is exactly 1.0"},
		{"tomato", "tomatoes", 7, 9, 6, 10, 0.6000, true, "the plural case from user story 8"},
		{"onion", "onions", 6, 7, 5, 8, 0.6250, true, "plural, worked in full in section 3"},
		{"breast", "breasts", 7, 8, 6, 9, 0.6667, true, "plural inside a multi-token term"},
		{"potato", "potatoes", 7, 9, 6, 10, 0.6000, true, "-o/-oes plural"},
		{"egg", "eggs", 4, 5, 3, 6, 0.5000, true, "short tokens still clear the bar"},
		{"oat", "oats", 4, 5, 3, 6, 0.5000, true, "three-letter token"},
		{"mushroom", "mushrooms", 9, 10, 8, 11, 0.7273, true, "long token, plural"},
		{"yogurt", "yoghurt", 7, 8, 5, 10, 0.5000, true, "spelling variant, not a plural"},
		{"egg", "eggplant", 4, 9, 3, 10, 0.3000, true, "boundary: exactly at threshold, and the comparison is >=, not >"},
		{"rice", "ricotta", 5, 8, 3, 10, 0.3000, true, "second boundary case, same reason"},
		{"chicken", "chickpea", 8, 9, 5, 12, 0.4167, true, "section 7.1: expected to be a non-match and is not"},
		{"chicken", "chickpeas", 8, 10, 5, 13, 0.3846, true, "same family, asserted so the behaviour is recorded"},
		{"chicken", "kitchen", 8, 8, 1, 15, 0.0667, false, "anagram-ish token that must not match"},
		{"chicken", "chili", 8, 6, 3, 11, 0.2727, false, "same first letters, below threshold"},
		{"tomato", "potato", 7, 7, 2, 12, 0.1667, false, "rhyming foods stay distinct"},
		{"basil", "basmati", 6, 8, 3, 11, 0.2727, false, "near-miss just under the bar"},
		{"lemon", "lime", 6, 5, 1, 10, 0.1000, false, "related foods, unrelated spellings"},
		{"beef", "broth", 5, 6, 1, 10, 0.1000, false, "words from the same recipe line"},
		{"kale", "cake", 5, 5, 0, 10, 0.0000, false, "same letters, no shared trigram"},
		{"chicken", "stock", 8, 6, 0, 14, 0.0000, false, "the other token of the false-positive pair scores zero"},
		{"yellow", "onion", 7, 6, 0, 13, 0.0000, false, "descriptor token contributes nothing on its own"},
		{"butter", "buttermilk", 7, 11, 6, 12, 0.5000, true, "prefix containment matches"},
		{"milk", "buttermilk", 5, 11, 3, 13, 0.2308, false, "suffix containment does not; the asymmetry must not be fixed"},
		{"bread", "breadcrumbs", 6, 12, 5, 13, 0.3846, true, "prefix containment again"},
		{"corn", "cornstarch", 5, 11, 4, 12, 0.3333, true, "prefix containment, near the bar"},
		{"apple", "pineapple", 6, 10, 4, 12, 0.3333, true, "compound food, accepted"},
		{"pepper", "pepperoni", 7, 10, 6, 11, 0.5455, true, "known false positive, prefix family"},
		{"pea", "peanut", 4, 7, 3, 8, 0.3750, true, "known false positive, short token"},
		{"celery", "celeriac", 7, 9, 5, 11, 0.4545, true, "known false positive"},
		{"parsley", "parsnip", 8, 8, 4, 12, 0.3333, true, "known false positive, just over the bar"},
		{"onion", "union", 6, 6, 3, 9, 0.3333, true, "not a food pair, but pins the arithmetic"},
		{"beef", "beets", 5, 6, 3, 8, 0.3750, true, "known false positive between real foods"},
	}

	for _, test := range tests {
		t.Run(test.a+" / "+test.b, func(t *testing.T) {
			if got := len(trigrams(test.a)); got != test.countA {
				t.Errorf("|trigrams(%q)| = %d, want %d", test.a, got, test.countA)
			}
			if got := len(trigrams(test.b)); got != test.countB {
				t.Errorf("|trigrams(%q)| = %d, want %d", test.b, got, test.countB)
			}

			intersection, union := trigrams(test.a).overlap(trigrams(test.b))
			if intersection != test.intersection || union != test.union {
				t.Errorf("similarity(%q, %q) = %d/%d, want %d/%d",
					test.a, test.b, intersection, union, test.intersection, test.union)
			}

			if got := round4(float64(intersection) / float64(union)); got != test.sim {
				t.Errorf("similarity(%q, %q) = %.4f, want %.4f", test.a, test.b, got, test.sim)
			}

			if got := trigrams(test.a).matches(trigrams(test.b)); got != test.match {
				t.Errorf("matches(%q, %q) = %v, want %v (%s)",
					test.a, test.b, got, test.match, test.protects)
			}
		})
	}
}

// Similarity is symmetric — |A ∩ B| and |A ∪ B| do not depend on argument
// order — so the fixture table only needs to state each pair once.
func TestSimilarityIsSymmetric(t *testing.T) {
	pairs := [][2]string{{"egg", "eggplant"}, {"milk", "buttermilk"}, {"chicken", "chickpea"}}
	for _, pair := range pairs {
		forward, forwardUnion := trigrams(pair[0]).overlap(trigrams(pair[1]))
		back, backUnion := trigrams(pair[1]).overlap(trigrams(pair[0]))
		if forward != back || forwardUnion != backUnion {
			t.Errorf("similarity(%q, %q) = %d/%d but reversed = %d/%d",
				pair[0], pair[1], forward, forwardUnion, back, backUnion)
		}
	}
}

// Section 3: trigrams are a set, so a token repeating one contributes it once.
// An implementation using a multiset scores similarity("banana", "banana")
// below 1.0 and disagrees with pg_trgm.
func TestDuplicateTrigramsCollapse(t *testing.T) {
	// "  banana " yields __b, _ba, ban, ana, nan, ana, na_ — ana twice.
	if got := len(trigrams("banana")); got != 6 {
		t.Fatalf("|trigrams(\"banana\")| = %d, want 6: duplicate trigrams collapse", got)
	}

	intersection, union := trigrams("banana").overlap(trigrams("banana"))
	if intersection != union {
		t.Fatalf("overlap(\"banana\", \"banana\") = %d/%d, want an exact 1.0", intersection, union)
	}
}

// Section 3's padding rule, spelled out: two leading spaces and one trailing,
// which is what makes a shared prefix worth more than a shared suffix.
func TestPaddingRule(t *testing.T) {
	got := trigrams("onion")
	if len(got) != 6 {
		t.Fatalf("|trigrams(\"onion\")| = %d, want 6", len(got))
	}
	for _, want := range []string{"  o", " on", "oni", "nio", "ion", "on "} {
		if _, ok := got[want]; !ok {
			t.Errorf("trigrams(\"onion\") is missing %q", want)
		}
	}
}

// The threshold is 0.3, and the comparison that applies it is integer
// arithmetic. 3/10 is exactly the threshold and must match; one trigram fewer
// in the intersection must not.
func TestThresholdComparisonIsExactAtTheBoundary(t *testing.T) {
	if MatchThreshold != 0.3 {
		t.Fatalf("MatchThreshold = %v, want 0.3", MatchThreshold)
	}
	if !clearsThreshold(3, 10) {
		t.Error("3/10 is exactly the threshold and the comparison is >=, so it must clear it")
	}
	if clearsThreshold(2, 10) {
		t.Error("2/10 is below the threshold")
	}
	if !clearsThreshold(1, 3) {
		t.Error("1/3 is above the threshold")
	}
	if clearsThreshold(0, 10) {
		t.Error("a pair sharing no trigram never clears the threshold")
	}
}

func round4(value float64) float64 {
	return math.Round(value*10000) / 10000
}
