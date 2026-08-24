package matching

import (
	"reflect"
	"testing"
)

// The scenarios below are the ranking fixtures of
// docs/ingredient-matching-algorithm.md section 6.3, adapted to the exact-token
// scoring this ticket implements: pantry items that the document matches by
// trigram similarity are written here in the form the Ingredient uses, so the
// pair still matches. Scenario A is unchanged, Scenario B drops the rows whose
// ordering depends on a fuzzy pair and uses `egg` for `eggs`, Scenario C is
// unchanged. When fuzzy scoring lands, these revert to the document's own
// inputs.
func candidate(id, name string, ingredients ...string) Candidate {
	return Candidate{RecipeID: id, Name: name, IngredientNames: ingredients}
}

func rankedIDs(matches []Match) []string {
	ids := make([]string, 0, len(matches))
	for _, match := range matches {
		ids = append(ids, match.RecipeID)
	}
	return ids
}

// Scenario A of docs/ingredient-matching-algorithm.md section 6.3: Coverage is
// the primary sort, so being close matters more than being small.
func TestRankOrdersByCoverageDescending(t *testing.T) {
	pantry := []string{"chicken", "onion", "garlic", "carrot", "celery", "salt", "pepper", "bay leaves", "noodles"}
	candidates := []Candidate{
		candidate("garlic-bread", "Garlic Bread", "garlic", "butter", "bread", "parsley"),
		candidate("chicken-noodle-soup", "Chicken Noodle Soup",
			"chicken", "onion", "carrot", "celery", "garlic", "salt",
			"pepper", "bay leaves", "noodles", "thyme", "parsley", "butter"),
	}

	matches := Rank(pantry, candidates)

	want := []string{"chicken-noodle-soup", "garlic-bread"}
	if got := rankedIDs(matches); !reflect.DeepEqual(got, want) {
		t.Fatalf("ranked %v, want %v", got, want)
	}
	soup := matches[0]
	if len(soup.Matched) != 9 {
		t.Fatalf("expected 9 matched Ingredient Terms, got %d (%v)", len(soup.Matched), soup.Matched)
	}
	wantMissing := []string{"thyme", "parsley", "butter"}
	if !reflect.DeepEqual(soup.Missing, wantMissing) {
		t.Fatalf("missing = %v, want %v in the Recipe's own Ingredient order", soup.Missing, wantMissing)
	}
	if soup.Coverage() != 0.75 {
		t.Fatalf("coverage = %v, want 0.75", soup.Coverage())
	}
}

// Scenario C: equal Coverage with unequal denominators falls to the Missing
// Ingredient count, and only then to the Recipe name.
func TestRankBreaksCoverageTiesOnMissingCount(t *testing.T) {
	pantry := []string{"rice", "onion", "egg"}
	candidates := []Candidate{
		candidate("fried-rice", "Fried Rice", "rice", "onion", "egg", "soy sauce", "oil", "peas"),
		candidate("onion-soup", "Onion Soup", "onion", "broth", "bread", "cheese", "butter", "thyme"),
		candidate("rice-pilaf", "Rice Pilaf", "rice", "onion", "broth", "butter"),
	}

	matches := Rank(pantry, candidates)

	want := []string{"rice-pilaf", "fried-rice", "onion-soup"}
	if got := rankedIDs(matches); !reflect.DeepEqual(got, want) {
		t.Fatalf("ranked %v, want %v", got, want)
	}
}

// Scenario B, restricted to the exact-token pairs this ticket scores: fully
// covered Recipes come first and tie on the Recipe name.
func TestRankBreaksRemainingTiesOnNameThenID(t *testing.T) {
	pantry := []string{"egg", "butter", "flour", "sugar", "milk"}
	candidates := []Candidate{
		candidate("crepes", "Crepes", "egg", "flour", "milk", "butter"),
		candidate("pound-cake", "Pound Cake", "butter", "sugar", "egg", "flour", "vanilla", "salt"),
		candidate("butter-cookies", "Butter Cookies", "butter", "sugar", "flour", "egg"),
	}

	matches := Rank(pantry, candidates)

	want := []string{"butter-cookies", "crepes", "pound-cake"}
	if got := rankedIDs(matches); !reflect.DeepEqual(got, want) {
		t.Fatalf("ranked %v, want %v", got, want)
	}
}

func TestRankBreaksNameTiesOnRecipeID(t *testing.T) {
	pantry := []string{"egg"}
	candidates := []Candidate{
		candidate("recipe-b", "Omelette", "egg"),
		candidate("recipe-a", "Omelette", "egg"),
	}

	matches := Rank(pantry, candidates)

	want := []string{"recipe-a", "recipe-b"}
	if got := rankedIDs(matches); !reflect.DeepEqual(got, want) {
		t.Fatalf("ranked %v, want %v", got, want)
	}
}

func TestRankMatchesOnAnySharedToken(t *testing.T) {
	pantry := []string{"chicken breast"}
	candidates := []Candidate{
		candidate("thighs", "Chicken Thighs", "chicken thighs"),
		candidate("basmati", "Basmati Rice", "basmati rice"),
	}

	matches := Rank(pantry, candidates)

	want := []string{"thighs"}
	if got := rankedIDs(matches); !reflect.DeepEqual(got, want) {
		t.Fatalf("ranked %v, want %v: a Recipe sharing no token is not a Match", got, want)
	}
	if matches[0].Coverage() != 1 {
		t.Fatalf("coverage = %v, want 1", matches[0].Coverage())
	}
}

func TestRankDiscardsIngredientsThatNormalizeToEmpty(t *testing.T) {
	pantry := []string{"salt"}
	candidates := []Candidate{
		candidate("seasoned", "Seasoned Something", "salt", "to taste", "freshly chopped"),
	}

	matches := Rank(pantry, candidates)

	if len(matches) != 1 {
		t.Fatalf("expected one match, got %d", len(matches))
	}
	if matches[0].Coverage() != 1 {
		t.Fatalf("coverage = %v, want 1: ingredients normalizing to empty are not counted", matches[0].Coverage())
	}
	if len(matches[0].Missing) != 0 {
		t.Fatalf("missing = %v, want none", matches[0].Missing)
	}
}

func TestRankDeduplicatesIngredientTerms(t *testing.T) {
	pantry := []string{"onions"}
	candidates := []Candidate{
		candidate("onions", "Onions", "2 large onions", "1 small onions", "butter"),
	}

	matches := Rank(pantry, candidates)

	if len(matches[0].Matched)+len(matches[0].Missing) != 2 {
		t.Fatalf("expected 2 Ingredient Terms, got matched %v missing %v", matches[0].Matched, matches[0].Missing)
	}
}

func TestRankDeduplicatesIngredientsThatDifferOnlyByDescriptor(t *testing.T) {
	pantry := []string{"chicken"}
	candidates := []Candidate{
		candidate("roast", "Roast Chicken", "Chicken breasts", "Boneless, skinless chicken breasts", "butter"),
	}

	matches := Rank(pantry, candidates)

	if got := len(matches[0].Matched); got != 1 {
		t.Fatalf("matched %d Ingredient Terms (%v), want 1: the two chicken lines normalize alike", got, matches[0].Matched)
	}
	if got := matches[0].Coverage(); got != 0.5 {
		t.Fatalf("coverage = %v, want 0.5", got)
	}
}

func TestRankDeduplicatesPantryItems(t *testing.T) {
	pantry := []string{"Onion", "onion", "", "   "}
	candidates := []Candidate{
		candidate("onion-soup", "Onion Soup", "onion", "broth"),
	}

	matches := Rank(pantry, candidates)

	if got := matches[0].Matched; !reflect.DeepEqual(got, []string{"onion"}) {
		t.Fatalf("matched = %v, want one Ingredient Term: Coverage counts Ingredient Terms, not Pantry Items", got)
	}
}

func TestRankExcludesRecipesWithNoIngredientTerms(t *testing.T) {
	pantry := []string{"salt"}
	candidates := []Candidate{
		candidate("noise", "Noise", "to taste", "freshly chopped"),
		candidate("empty", "Empty"),
	}

	matches := Rank(pantry, candidates)

	if len(matches) != 0 {
		t.Fatalf("expected no matches, got %v", rankedIDs(matches))
	}
}

func TestRankWithAPantryThatNormalizesToNothing(t *testing.T) {
	pantry := []string{"to taste", "   ", "freshly chopped"}
	candidates := []Candidate{
		candidate("soup", "Soup", "onion", "broth"),
	}

	if matches := Rank(pantry, candidates); len(matches) != 0 {
		t.Fatalf("expected no matches, got %v", rankedIDs(matches))
	}
}
