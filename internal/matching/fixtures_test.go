package matching

import (
	"reflect"
	"testing"
)

// Section 6.2 of docs/ingredient-matching-algorithm.md, transcribed row for
// row. The document calls this table the contract between the Go and Swift
// implementations, so it is consumed here as a table rather than paraphrased
// into prose assertions.
//
// Each row asserts two things. The normalized pair is asserted for every row —
// that column is normalization's output and is entirely settled by this
// ticket. The verdict is asserted only for the rows exact-token scoring can
// decide; the four rows whose verdict rests on trigram similarity carry
// needsSimilarity and have their verdict deferred to the fuzzy-matching
// ticket, which will drop the flag rather than rewrite the row.
func TestSection62Fixtures(t *testing.T) {
	tests := []struct {
		pantryItem       string
		ingredientName   string
		pantryTokens     []string
		ingredientTokens []string
		match            bool
		needsSimilarity  bool
		protects         string
	}{
		{"tomato", "tomato", []string{"tomato"}, []string{"tomato"}, true, false, "exact match"},
		{"Tomato", "tomato", []string{"tomato"}, []string{"tomato"}, true, false, "casing"},
		{"  ToMaTo  ", "Tomato", []string{"tomato"}, []string{"tomato"}, true, false, "surrounding whitespace and mixed case"},
		{"tomato", "Tomatoes", []string{"tomato"}, []string{"tomatoes"}, true, true, "plural, sim 0.6"},
		{"Tomatoes", "1 tomato, diced", []string{"tomatoes"}, []string{"tomato"}, true, true, "quantity and descriptor stripped from the Ingredient"},
		{"chicken breast", "boneless skinless chicken breasts", []string{"breast", "chicken"}, []string{"breasts", "chicken"}, true, false, "the motivating case"},
		{"chicken breast", "Boneless, Skinless Chicken Breasts", []string{"breast", "chicken"}, []string{"breasts", "chicken"}, true, false, "the motivating case, punctuated"},
		{"2 Yellow Onions", "onion", []string{"onions", "yellow"}, []string{"onion"}, true, true, "quantity-prefixed input"},
		{"1 1/2 cups all-purpose flour", "Flour", []string{"all", "flour", "purpose"}, []string{"flour"}, true, false, "fraction, measurement stopword, hyphen split"},
		{"½ cup ⅓ milk", "Milk", []string{"milk"}, []string{"milk"}, true, false, "unicode fractions are boundaries, not digits"},
		{"500g Beef", "ground beef", []string{"beef"}, []string{"beef"}, true, false, "digit-fused unit, and a descriptor on the Ingredient"},
		{"chicken breast", "chicken stock", []string{"breast", "chicken"}, []string{"chicken", "stock"}, true, false, "the known false positive, asserted deliberately"},
		{"heavy cream", "sour cream", []string{"cream", "heavy"}, []string{"cream", "sour"}, true, false, "shared head token decides it"},
		{"olive oil", "vegetable oil", []string{"oil", "olive"}, []string{"oil", "vegetable"}, true, false, "same again"},
		{"chicken", "chickpeas", []string{"chicken"}, []string{"chickpeas"}, true, true, "expected to be a non-match; is not"},
		{"chicken", "chicken thighs", []string{"chicken"}, []string{"chicken", "thighs"}, true, false, "one token against a multi-token Term"},
		{"tomato", "potato", []string{"tomato"}, []string{"potato"}, false, false, "near-miss that must not match"},
		{"basil", "basmati rice", []string{"basil"}, []string{"basmati", "rice"}, false, false, "both token pairs below the bar"},
		{"beef", "chicken broth", []string{"beef"}, []string{"broth", "chicken"}, false, false, "no token pair clears the bar"},
		{"", "tomato", nil, []string{"tomato"}, false, false, "empty input matches nothing"},
		{"   ", "tomato", nil, []string{"tomato"}, false, false, "whitespace-only input matches nothing"},
		{",.-/()", "tomato", nil, []string{"tomato"}, false, false, "punctuation-only input normalizes to empty"},
		{"2 1/2", "tomato", nil, []string{"tomato"}, false, false, "digits-only input normalizes to empty"},
		{"freshly chopped", "fresh chopped tomato", nil, []string{"tomato"}, false, false, "an all-stopword Pantry Item matches nothing, even sharing those stopwords"},
		{"to taste", "salt", nil, []string{"salt"}, false, false, "measurement stopwords alone normalize to empty"},
		{"tomato", "to taste", []string{"tomato"}, nil, false, false, "an Ingredient normalizing to empty is discarded entirely"},
		{"Salt & Pepper", "black pepper", []string{"pepper", "salt"}, []string{"black", "pepper"}, true, false, "ampersand as a boundary; second token carries the match"},
	}

	for _, test := range tests {
		t.Run(test.pantryItem+" / "+test.ingredientName, func(t *testing.T) {
			if got := Normalize(test.pantryItem).Tokens(); !sameTokens(got, test.pantryTokens) {
				t.Errorf("Normalize(%q) = %v, want %v", test.pantryItem, got, test.pantryTokens)
			}
			if got := Normalize(test.ingredientName).Tokens(); !sameTokens(got, test.ingredientTokens) {
				t.Errorf("Normalize(%q) = %v, want %v", test.ingredientName, got, test.ingredientTokens)
			}

			if test.needsSimilarity {
				t.Skipf("document's verdict is %v, but it rests on trigram similarity: %s",
					verdict(test.match), test.protects)
			}

			got := matchesFixture(test.pantryItem, test.ingredientName)
			if got != test.match {
				t.Errorf("matches(%q, %q) = %v, want %v (%s)",
					test.pantryItem, test.ingredientName, got, test.match, test.protects)
			}
		})
	}
}

// matchesFixture asks the production path — Rank, and so pantry.score — whether
// one Pantry Item matches one Ingredient. Going through Rank rather than
// reaching for pantry.holds keeps section 5's rule that an Ingredient
// normalizing to empty is discarded from the Recipe's Ingredient Term set in
// the one place that owns it, so a fixture row cannot pass against a
// reimplementation of a rule score has since diverged from.
func matchesFixture(pantryItem, ingredientName string) bool {
	matches := Rank([]string{pantryItem}, []Candidate{{
		RecipeID:        "fixture",
		Name:            "Fixture",
		IngredientNames: []string{ingredientName},
	}})

	return len(matches) == 1 && len(matches[0].Matched) == 1
}

func verdict(match bool) string {
	if match {
		return "MATCH"
	}
	return "no match"
}

// Colour and variety words are deliberately not stopwords: dropping them would
// erase `sweet potato` and `green onion` as distinct foods.
func TestColourAndVarietyWordsSurviveNormalization(t *testing.T) {
	tests := []struct {
		raw    string
		tokens []string
	}{
		{"sweet potato", []string{"potato", "sweet"}},
		{"green onion", []string{"green", "onion"}},
		{"red bell pepper", []string{"bell", "pepper", "red"}},
		{"yellow onion", []string{"onion", "yellow"}},
		{"baby spinach", []string{"baby", "spinach"}},
		{"wild rice", []string{"rice", "wild"}},
		{"salted butter", []string{"butter", "salted"}},
		{"sweetened condensed milk", []string{"condensed", "milk", "sweetened"}},
		{"heavy cream", []string{"cream", "heavy"}},
		{"chicken stock", []string{"chicken", "stock"}},
		{"beef broth", []string{"beef", "broth"}},
		{"fat free yogurt", []string{"fat", "free", "yogurt"}},

		// The document's one named loss: `whole` is a MeasurementUnit.piece
		// alias, so it goes with the measurements. `flour` still matches.
		{"whole wheat flour", []string{"flour", "wheat"}},
	}

	for _, test := range tests {
		t.Run(test.raw, func(t *testing.T) {
			if got := Normalize(test.raw).Tokens(); !sameTokens(got, test.tokens) {
				t.Fatalf("Normalize(%q) = %v, want %v", test.raw, got, test.tokens)
			}
		})
	}
}

// Sweet potato must stay distinct from potato, which is what keeping the
// colour and variety words buys.
func TestVarietyWordsDoNotCollapseDistinctFoods(t *testing.T) {
	sweetPotato := Normalize("sweet potato")
	if len(sweetPotato.Tokens()) != 2 {
		t.Fatalf("sweet potato = %v, want both tokens", sweetPotato.Tokens())
	}
	if sweetPotato.key() == Normalize("potato").key() {
		t.Fatal("sweet potato and potato must not be the same Ingredient Term")
	}
	if Normalize("green onion").key() == Normalize("onion").key() {
		t.Fatal("green onion and onion must not be the same Ingredient Term")
	}
}

// Every token of both stopword lists drops, and a string made only of them
// normalizes to the empty Term rather than crashing or surviving as noise.
func TestEveryStopwordDrops(t *testing.T) {
	for _, list := range []struct {
		name  string
		words map[string]struct{}
	}{
		{"MeasurementStopwords", MeasurementStopwords},
		{"DescriptorStopwords", DescriptorStopwords},
	} {
		for word := range list.words {
			if got := Normalize(word); !got.IsEmpty() {
				t.Errorf("%s: Normalize(%q) = %v, want the empty Term", list.name, word, got.Tokens())
			}
			if got := Normalize("tomato " + word); !sameTokens(got.Tokens(), []string{"tomato"}) {
				t.Errorf("%s: Normalize(%q) = %v, want just the food token", list.name, "tomato "+word, got.Tokens())
			}
		}
	}
}

// sameTokens compares two token sets, treating a nil result and an empty
// expectation as equal: normalization returns no tokens rather than an empty
// slice, and the empty Term is a legal outcome throughout these tables.
func sameTokens(got, want []string) bool {
	if len(got) == 0 && len(want) == 0 {
		return true
	}
	return reflect.DeepEqual(got, want)
}
