package matching

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

// A rankingFixture is one scenario of section 6.3 of
// docs/ingredient-matching-algorithm.md: a pantry, and the Recipes it is
// scored against.
type rankingFixture struct {
	scenario string
	pantry   []string
	recipes  []recipeFixture
}

// A recipeFixture is one row of a scenario's table. Rows are held in the
// document's own order, which is not the order Rank returns them in; rank
// carries that.
type recipeFixture struct {
	id          string
	name        string
	ingredients []string
	terms       int
	matched     int
	missing     []string
	coverage    float64
	rank        int
}

// Section 6.3, transcribed scenario for scenario. Every column is asserted,
// not just the order: an order that comes out right for the wrong reason — the
// wrong Ingredient Term count, a Missing Ingredient counted as matched — still
// disagrees with the Swift implementation on the next corpus.
var section63 = []rankingFixture{
	{
		scenario: "A: Coverage is the primary sort",
		pantry:   []string{"chicken", "onion", "garlic", "carrot", "celery", "salt", "pepper", "bay leaves", "noodles"},
		recipes: []recipeFixture{
			{
				id:   "chicken-noodle-soup",
				name: "Chicken Noodle Soup",
				ingredients: []string{
					"chicken", "onion", "carrot", "celery", "garlic", "salt",
					"pepper", "bay leaves", "noodles", "thyme", "parsley", "butter",
				},
				terms:    12,
				matched:  9,
				missing:  []string{"thyme", "parsley", "butter"},
				coverage: 0.7500,
				rank:     1,
			},
			{
				id:          "garlic-bread",
				name:        "Garlic Bread",
				ingredients: []string{"garlic", "butter", "bread", "parsley"},
				terms:       4,
				matched:     1,
				missing:     []string{"butter", "bread", "parsley"},
				coverage:    0.2500,
				rank:        2,
			},
		},
	},
	{
		// The document names this scenario's two tied pairs as the ones that
		// reach the name tie-break, so their ids run counter to their names: a
		// run that skipped the name and fell through to the id returns each
		// pair the other way round.
		scenario: "B: full Coverage first, then the tie-breaks",
		pantry:   []string{"eggs", "butter", "flour", "sugar", "milk"},
		recipes: []recipeFixture{
			{
				id:          "b-butter-cookies",
				name:        "Butter Cookies",
				ingredients: []string{"butter", "sugar", "flour", "egg"},
				terms:       4,
				matched:     4,
				coverage:    1.0000,
				rank:        1,
			},
			{
				id:          "a-crepes",
				name:        "Crepes",
				ingredients: []string{"eggs", "flour", "milk", "butter"},
				terms:       4,
				matched:     4,
				coverage:    1.0000,
				rank:        2,
			},
			{
				id:          "pound-cake",
				name:        "Pound Cake",
				ingredients: []string{"butter", "sugar", "eggs", "flour", "vanilla", "salt"},
				terms:       6,
				matched:     4,
				missing:     []string{"vanilla", "salt"},
				coverage:    0.6667,
				rank:        3,
			},
			{
				id:          "b-pancakes",
				name:        "Pancakes",
				ingredients: []string{"flour", "sugar", "eggs", "milk", "butter", "baking powder", "salt", "oil"},
				terms:       8,
				matched:     5,
				missing:     []string{"baking powder", "salt", "oil"},
				coverage:    0.6250,
				rank:        4,
			},
			{
				id:          "a-waffles",
				name:        "Waffles",
				ingredients: []string{"flour", "eggs", "milk", "butter", "sugar", "baking soda", "salt", "cinnamon"},
				terms:       8,
				matched:     5,
				missing:     []string{"baking soda", "salt", "cinnamon"},
				coverage:    0.6250,
				rank:        5,
			},
		},
	},
	{
		scenario: "C: equal Coverage, unequal denominators",
		pantry:   []string{"rice", "onion", "egg"},
		recipes: []recipeFixture{
			{
				id:          "rice-pilaf",
				name:        "Rice Pilaf",
				ingredients: []string{"rice", "onion", "broth", "butter"},
				terms:       4,
				matched:     2,
				missing:     []string{"broth", "butter"},
				coverage:    0.5000,
				rank:        1,
			},
			{
				id:          "fried-rice",
				name:        "Fried Rice",
				ingredients: []string{"rice", "onion", "egg", "soy sauce", "oil", "peas"},
				terms:       6,
				matched:     3,
				missing:     []string{"soy sauce", "oil", "peas"},
				coverage:    0.5000,
				rank:        2,
			},
			{
				id:          "onion-soup",
				name:        "Onion Soup",
				ingredients: []string{"onion", "broth", "bread", "cheese", "butter", "thyme"},
				terms:       6,
				matched:     1,
				missing:     []string{"broth", "bread", "cheese", "butter", "thyme"},
				coverage:    0.1667,
				rank:        3,
			},
		},
	},
}

func TestSection63Fixtures(t *testing.T) {
	for _, fixture := range section63 {
		t.Run(fixture.scenario, func(t *testing.T) {
			matches := Rank(fixture.pantry, fixture.candidates())

			if got, want := rankedIDs(matches), fixture.expectedOrder(); !reflect.DeepEqual(got, want) {
				t.Fatalf("ranked %v, want %v", got, want)
			}

			for _, recipe := range fixture.recipes {
				match := matches[recipe.rank-1]

				if got := len(match.Matched) + len(match.Missing); got != recipe.terms {
					t.Errorf("%s: %d Ingredient Terms, want %d", recipe.name, got, recipe.terms)
				}
				if got := len(match.Matched); got != recipe.matched {
					t.Errorf("%s: matched %d Ingredient Terms (%v), want %d",
						recipe.name, got, match.Matched, recipe.matched)
				}
				if !sameTokens(match.Missing, recipe.missing) {
					t.Errorf("%s: missing = %v, want %v in the Recipe's own Ingredient order",
						recipe.name, match.Missing, recipe.missing)
				}
				if got := round4(match.Coverage()); got != recipe.coverage {
					t.Errorf("%s: coverage = %.4f, want %.4f", recipe.name, got, recipe.coverage)
				}
			}
		})
	}
}

// Scenario A's other assertion: the Pantry Item `bay leaves` normalizes to
// `{bay, leaves}` and covers a Recipe's one-token `bay` Term on the first of
// them. The two strings are deliberately not the same string — comparing the
// Pantry Item whole would miss this, which is the point of comparing tokens.
func TestScenarioAMatchesBayLeavesAgainstBay(t *testing.T) {
	matches := Rank([]string{"bay leaves"}, []Candidate{
		candidate("soup", "Soup", "bay", "noodles"),
	})

	if got := matches[0].Matched; !reflect.DeepEqual(got, []string{"bay"}) {
		t.Fatalf("matched = %v, want the Recipe's bay Ingredient", got)
	}
	if got := matches[0].Missing; !reflect.DeepEqual(got, []string{"noodles"}) {
		t.Fatalf("missing = %v, want noodles: `leaves` covers nothing on its own", got)
	}
}

// Scenario D, minus its first row: a pantry of nothing but blanks is rejected
// before anything is scored, so that row is asserted where the rejection lives,
// in the handler tests.
func TestSection63ScenarioD(t *testing.T) {
	t.Run("pantry items that all normalize to empty", func(t *testing.T) {
		matches := Rank([]string{"to taste", "freshly chopped"}, []Candidate{
			candidate("soup", "Soup", "onion", "broth"),
		})

		if len(matches) != 0 {
			t.Fatalf("expected no Matches, got %v", rankedIDs(matches))
		}
	})

	t.Run("duplicates and blanks in the pantry", func(t *testing.T) {
		matches := Rank([]string{"Onion", "onion", "", " onions "}, []Candidate{
			candidate("onion-soup", "Onion Soup", "onion", "broth"),
		})

		// `onion` and `onions` are two distinct Ingredient Terms, and both
		// match the Recipe's one `onion` Term. That is one matched Ingredient
		// Term: Coverage counts Ingredient Terms, never Pantry Items.
		if got := matches[0].Matched; !reflect.DeepEqual(got, []string{"onion"}) {
			t.Fatalf("matched = %v, want one Ingredient Term", got)
		}
		if got := matches[0].Coverage(); got != 0.5 {
			t.Fatalf("coverage = %v, want 0.5", got)
		}
	})

	t.Run("recipe with every ingredient normalizing to empty", func(t *testing.T) {
		matches := Rank([]string{"salt"}, []Candidate{
			candidate("noise", "Noise", "to taste", "freshly chopped"),
			candidate("empty", "Empty"),
		})

		if len(matches) != 0 {
			t.Fatalf("expected no Matches, got %v: zero Ingredient Terms is Coverage 0", rankedIDs(matches))
		}
	})
}

// The ranking property behind user story 7, over every pair of Recipe sizes
// rather than the document's single pair: at an equal Missing Ingredient count
// the larger Recipe is the better Match, because Coverage decides first.
//
// The names and ids run backwards against the expected order, so a run that
// reached either of those tie-breaks would come out reversed.
func TestRankPrefersTheLargerRecipeAtAnEqualMissingCount(t *testing.T) {
	const missingCount = 3

	covered := []string{"rice", "onion", "egg", "butter", "flour", "sugar", "milk", "thyme", "bread"}
	absent := []string{"cheese", "lemon", "kale"}
	assertIndependent(t, append(append([]string{}, covered...), absent...))

	candidates := make([]Candidate, 0, len(covered))
	want := make([]string, 0, len(covered))
	for size := len(covered) + missingCount; size > missingCount; size-- {
		// Bigger Recipes get later ids and later names, the reverse of the
		// order Coverage puts them in.
		id := fmt.Sprintf("recipe-%02d", size)
		candidates = append(candidates, candidate(id, id,
			append(append([]string{}, covered[:size-missingCount]...), absent...)...))
		want = append(want, id)
	}

	if got := rankedIDs(Rank(covered, candidates)); !reflect.DeepEqual(got, want) {
		t.Fatalf("ranked %v, want %v: at three missing each, the larger Recipe covers more", got, want)
	}
}

// Section 5's third tie-break is UTF-8 byte order, not whatever collation a
// language hands you. Two names differing only in how an accent is composed
// are canonically equivalent, so Swift's `<` calls them equal and falls
// through to the Recipe id; the document pins the comparison to the bytes so
// that both implementations reach the same answer. The ids below are arranged
// so a run that fell through would come out reversed.
func TestRankBreaksNameTiesByUTF8ByteOrder(t *testing.T) {
	const (
		composed   = "Crème Brûlée"                   // è, û and é precomposed
		decomposed = "Cre\u0300me Bru\u0302le\u0301e" // the same name, accents combining
	)

	tests := []struct {
		name       string
		candidates []Candidate
		want       []string
		protects   string
	}{
		{
			name: "accent composition",
			candidates: []Candidate{
				candidate("a-composed", composed, "egg"),
				candidate("b-decomposed", decomposed, "egg"),
			},
			want:     []string{"b-decomposed", "a-composed"},
			protects: "the combining form's plain `e` is byte 0x65, below the 0xc3 that opens the precomposed `è`",
		},
		{
			name: "ascii against an accented letter",
			candidates: []Candidate{
				candidate("a-eclair", "Éclair", "egg"),
				candidate("b-zucchini", "Zucchini Bread", "egg"),
			},
			want:     []string{"b-zucchini", "a-eclair"},
			protects: "byte order puts every ASCII letter below `É`, where a locale collation files it under E",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := rankedIDs(Rank([]string{"egg"}, test.candidates)); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("ranked %v, want %v: %s", got, test.want, test.protects)
			}
		})
	}
}

// Coverage is compared by cross-multiplication, so two Matches a hair apart
// stay a hair apart. The counts below cannot come from a real Recipe; they are
// here because dividing them in float64 collapses both ratios onto the same
// value, which is the only way to make the difference between the two
// comparisons observable from a test. An implementation that divides reads
// these two as tied, falls through to the Missing Ingredient count, and
// returns them the other way round.
func TestRankKeyComparesCoverageWithoutDividing(t *testing.T) {
	const (
		matched = 1<<54 + 1 // one Ingredient Term more than a third of...
		total   = 3 << 54   // ...a Recipe this size
	)

	wider := rankKey{matched: matched, missing: total - matched, name: "Wider", recipeID: "wider"}
	narrower := rankKey{matched: 1, missing: 2, name: "Narrower", recipeID: "narrower"}

	if float64(wider.matched)/float64(wider.total()) != float64(narrower.matched)/float64(narrower.total()) {
		t.Fatal("these counts no longer land on one float64, so this test no longer proves anything")
	}

	if !wider.precedes(narrower) {
		t.Error("the wider Recipe covers the larger fraction and must rank first")
	}
	if narrower.precedes(wider) {
		t.Error("the narrower Recipe covers the smaller fraction and must not rank first")
	}
}

// The same pantry over the same corpus produces the same order, whatever order
// the candidates arrive in and however many times they are scored. The final
// tie-break on Recipe id is what makes that true: without it, two Matches equal
// on every other key come back in whatever order the sort happened to leave
// them.
func TestRankIsDeterministicAcrossRunsAndInputOrder(t *testing.T) {
	pantry := []string{"eggs", "butter", "flour", "sugar", "milk", "rice", "onion"}

	candidates := make([]Candidate, 0)
	for _, fixture := range section63 {
		candidates = append(candidates, fixture.candidates()...)
	}

	// Two Recipes sharing a name and a Coverage, so the run has to reach the
	// Recipe id rather than stopping at the name.
	candidates = append(candidates,
		candidate("omelette-b", "Omelette", "eggs", "butter"),
		candidate("omelette-a", "Omelette", "eggs", "butter"),
	)

	first := Rank(pantry, candidates)
	if len(first) == 0 {
		t.Fatal("expected the corpus to produce Matches")
	}

	for run := 0; run < 50; run++ {
		if got := Rank(shuffled(run, pantry), shuffled(run, candidates)); !reflect.DeepEqual(got, first) {
			t.Fatalf("run %d ranked %v, want %v", run, rankedIDs(got), rankedIDs(first))
		}
	}
}

// PantryTokens is what narrows the corpus, so it is on the same hook: the same
// Pantry Items in a different order must narrow on the same tokens, or two
// identical requests can be scored against two different candidate sets.
func TestPantryTokensAreStableWhateverOrderTheItemsArriveIn(t *testing.T) {
	items := []string{"Onion", "chicken breast", "  ToMaTo  ", "onion", "1 1/2 cups all-purpose flour"}
	want := PantryTokens(items)

	for run := 0; run < 20; run++ {
		if got := PantryTokens(shuffled(run, items)); !reflect.DeepEqual(got, want) {
			t.Fatalf("run %d = %v, want %v", run, got, want)
		}
	}
}

func (f rankingFixture) candidates() []Candidate {
	candidates := make([]Candidate, 0, len(f.recipes))
	for _, recipe := range f.recipes {
		candidates = append(candidates, candidate(recipe.id, recipe.name, recipe.ingredients...))
	}
	return candidates
}

func (f rankingFixture) expectedOrder() []string {
	order := make([]string, len(f.recipes))
	for _, recipe := range f.recipes {
		order[recipe.rank-1] = recipe.id
	}
	return order
}

// shuffled returns a reordered copy of a slice, from a seed the caller names so
// a failing run can be reproduced. It stands in for the order the corpus and
// the pantry happen to arrive in, which nothing upstream promises.
func shuffled[T any](seed int, items []T) []T {
	reordered := append([]T{}, items...)
	rand.New(rand.NewSource(int64(seed))).Shuffle(len(reordered), func(i, j int) {
		reordered[i], reordered[j] = reordered[j], reordered[i]
	})
	return reordered
}

// assertIndependent fails unless no two of the given foods match each other,
// which is what lets a test hold one of them covered and another missing.
// Trigram similarity is generous enough that picking food names by eye is not
// good enough.
func assertIndependent(t *testing.T, foods []string) {
	t.Helper()

	for i, food := range foods {
		for _, other := range foods[i+1:] {
			if trigrams(food).matches(trigrams(other)) {
				t.Fatalf("%q and %q match each other, so this fixture proves nothing", food, other)
			}
		}
	}
}
