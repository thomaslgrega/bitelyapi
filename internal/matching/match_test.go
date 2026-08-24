package matching

import (
	"reflect"
	"testing"
)

// The rules of docs/ingredient-matching-algorithm.md section 5 that a whole
// scenario would only exercise in passing. The document's own ranking
// scenarios are transcribed in ranking_test.go.
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

// Section 5: the match is binary. A token pair at 1.0 and a token pair at
// exactly the threshold each make their Ingredient Term matched and nothing
// more, so Coverage stays a ratio the interface can read out loud.
func TestRankGivesNoPartialCreditForAWeakMatch(t *testing.T) {
	// egg/egg is 1.0; egg/eggplant is exactly 0.3, the boundary row of section
	// 6.1. Fractional credit would put Coverage somewhere below 2/3.
	pantry := []string{"egg"}
	candidates := []Candidate{
		candidate("shakshuka", "Shakshuka", "egg", "eggplant", "tomato"),
	}

	matches := Rank(pantry, candidates)

	if got := matches[0].Matched; !reflect.DeepEqual(got, []string{"egg", "eggplant"}) {
		t.Fatalf("matched = %v, want both the exact and the boundary pair", got)
	}
	if got := matches[0].Coverage(); got != 2.0/3.0 {
		t.Fatalf("coverage = %v, want exactly 2/3: a weak match counts wholly, not partly", got)
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
