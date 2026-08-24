package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/thomaslgrega/bitelyapi/internal/models"
)

func matchRequest(t *testing.T, h *RecipeHandler, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/recipes/match", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	h.MatchRecipes(rec, req)
	return rec
}

func decodeMatches(t *testing.T, rec *httptest.ResponseRecorder) []models.RecipeMatch {
	t.Helper()

	var matches []models.RecipeMatch
	if err := json.Unmarshal(rec.Body.Bytes(), &matches); err != nil {
		t.Fatalf("failed to decode matches: %v", err)
	}
	return matches
}

// Scenario D of docs/ingredient-matching-algorithm.md section 6.3 opens with
// this row. The rest of that scenario is scored rather than rejected, so it is
// asserted in the matching package.
func TestRecipeHandlerMatchRecipesRejectsAnEmptyPantry(t *testing.T) {
	tests := map[string]string{
		"invalid json": `{`,
		"empty list":   `[]`,
		// Nothing but blanks carries no Pantry Item, so it is rejected like an
		// empty list rather than answered with an empty one. That split is
		// issue 5's, not the algorithm document's; ADR-0003 records why it won.
		"all blank":  `["", "   "]`,
		"not a list": `{"pantry": ["onion"]}`,
	}

	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			h := NewRecipeHandler(fakeRecipeRepo{})

			rec := matchRequest(t, h, body)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
			}
		})
	}
}

// The other half of ADR-0003's split. A blank alongside a stopword-only string
// does not drag the request back into the 400 above: the blank is discarded and
// the stopword string is what the emptiness check sees.
func TestRecipeHandlerMatchRecipesWithAPantryThatNormalizesToNothing(t *testing.T) {
	bodies := map[string]string{
		"only stopword strings": `["to taste", "freshly chopped"]`,
		"a blank beside one":    `["", "to taste"]`,
	}

	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			h := NewRecipeHandler(fakeRecipeRepo{
				getMatchCandidatesFunc: func(ctx context.Context, tokens []string, limit int) ([]models.MatchCandidate, error) {
					t.Fatal("expected no repository call for a pantry with no Ingredient Terms")
					return nil, nil
				},
			})

			rec := matchRequest(t, h, body)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
			}
			if matches := decodeMatches(t, rec); len(matches) != 0 {
				t.Fatalf("expected an empty list, got %v", matches)
			}
			if body := rec.Body.String(); body != "[]\n" {
				t.Fatalf("expected an empty JSON list, got %q", body)
			}
		})
	}
}

// A Pantry Item that is nothing but measurements and descriptors contributes
// no tokens, so it neither widens narrowing nor drags the whole corpus in
// alongside the items that do name a food.
func TestRecipeHandlerMatchRecipesDropsPantryItemsThatAreAllStopwords(t *testing.T) {
	var gotTokens []string
	h := NewRecipeHandler(fakeRecipeRepo{
		getMatchCandidatesFunc: func(ctx context.Context, tokens []string, limit int) ([]models.MatchCandidate, error) {
			gotTokens = tokens
			return nil, nil
		},
	})

	rec := matchRequest(t, h, `["onion", "to taste", "freshly chopped", "1 1/2 cups"]`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if want := []string{"onion"}; !reflect.DeepEqual(gotTokens, want) {
		t.Fatalf("narrowed on %v, want %v", gotTokens, want)
	}
}

func TestRecipeHandlerMatchRecipesReturnsRepositoryError(t *testing.T) {
	h := NewRecipeHandler(fakeRecipeRepo{
		getMatchCandidatesFunc: func(ctx context.Context, tokens []string, limit int) ([]models.MatchCandidate, error) {
			return nil, errors.New("boom")
		},
	})

	rec := matchRequest(t, h, `["onion"]`)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestRecipeHandlerMatchRecipesDiscardsBlanksAndDuplicates(t *testing.T) {
	var gotTokens []string
	var gotLimit int
	h := NewRecipeHandler(fakeRecipeRepo{
		getMatchCandidatesFunc: func(ctx context.Context, tokens []string, limit int) ([]models.MatchCandidate, error) {
			gotTokens = tokens
			gotLimit = limit
			return nil, nil
		},
	})

	rec := matchRequest(t, h, `["Onion", "onion", "", "  ", " ONION ", "2 cups rice"]`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if want := []string{"onion", "rice"}; !reflect.DeepEqual(gotTokens, want) {
		t.Fatalf("narrowed on %v, want %v", gotTokens, want)
	}
	if gotLimit != candidateCeiling {
		t.Fatalf("narrowed with limit %d, want the candidate ceiling %d", gotLimit, candidateCeiling)
	}
}

func TestRecipeHandlerMatchRecipesWithAPantryThatMatchesNothing(t *testing.T) {
	h := NewRecipeHandler(fakeRecipeRepo{
		getMatchCandidatesFunc: func(ctx context.Context, tokens []string, limit int) ([]models.MatchCandidate, error) {
			// Narrowing is looser than scoring, so it can hand back a Recipe
			// that then covers nothing. `beef` against `chicken broth` is the
			// section 6.2 row where no token pair clears the threshold.
			return []models.MatchCandidate{
				{
					Recipe:      models.RecipeSummary{ID: "broth", Name: "Chicken Broth"},
					Ingredients: []string{"chicken", "broth"},
				},
			}, nil
		},
	})

	rec := matchRequest(t, h, `["beef"]`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if body := rec.Body.String(); body != "[]\n" {
		t.Fatalf("expected an empty JSON list, got %q", body)
	}
}

func TestRecipeHandlerMatchRecipesReturnsRankedMatches(t *testing.T) {
	h := NewRecipeHandler(fakeRecipeRepo{
		getMatchCandidatesFunc: func(ctx context.Context, tokens []string, limit int) ([]models.MatchCandidate, error) {
			return []models.MatchCandidate{
				{
					Recipe:      models.RecipeSummary{ID: "garlic-bread", Name: "Garlic Bread", Category: "sides", ThumbnailUrl: "https://example.test/bread.jpg", Calories: 200, TotalCookTime: 15},
					Ingredients: []string{"garlic", "butter", "bread"},
				},
				{
					Recipe:      models.RecipeSummary{ID: "garlic-rice", Name: "Garlic Rice", Category: "sides", Calories: 300, TotalCookTime: 20},
					Ingredients: []string{"rice", "garlic", "butter"},
				},
			}, nil
		},
	})

	rec := matchRequest(t, h, `["rice", "garlic"]`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if contentType := rec.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected a JSON content type, got %q", contentType)
	}

	matches := decodeMatches(t, rec)
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}

	best := matches[0]
	if best.ID != "garlic-rice" {
		t.Fatalf("expected the best Coverage first, got %q", best.ID)
	}
	if best.Name != "Garlic Rice" || best.Category != "sides" || best.Calories != 300 || best.TotalCookTime != 20 {
		t.Fatalf("expected the recipe card fields to be carried, got %+v", best)
	}
	if want := []string{"rice", "garlic"}; !reflect.DeepEqual(best.MatchedIngredients, want) {
		t.Fatalf("matched = %v, want %v", best.MatchedIngredients, want)
	}
	if want := []string{"butter"}; !reflect.DeepEqual(best.MissingIngredients, want) {
		t.Fatalf("missing = %v, want %v", best.MissingIngredients, want)
	}
	if best.Coverage <= matches[1].Coverage {
		t.Fatalf("expected Coverage descending, got %v then %v", best.Coverage, matches[1].Coverage)
	}
}

func TestRecipeHandlerMatchRecipesCapsTheResponse(t *testing.T) {
	h := NewRecipeHandler(fakeRecipeRepo{
		getMatchCandidatesFunc: func(ctx context.Context, tokens []string, limit int) ([]models.MatchCandidate, error) {
			candidates := make([]models.MatchCandidate, 0, 120)
			for i := range 120 {
				candidates = append(candidates, models.MatchCandidate{
					Recipe:      models.RecipeSummary{ID: fmt.Sprintf("recipe-%03d", i), Name: fmt.Sprintf("Recipe %03d", i)},
					Ingredients: []string{"onion"},
				})
			}
			return candidates, nil
		},
	})

	rec := matchRequest(t, h, `["onion"]`)

	matches := decodeMatches(t, rec)
	if len(matches) != 50 {
		t.Fatalf("expected the response capped at 50, got %d", len(matches))
	}
	if matches[0].ID != "recipe-000" {
		t.Fatalf("expected the cap to keep the best matches, got %q first", matches[0].ID)
	}
	if matches[0].MissingIngredients == nil {
		t.Fatal("expected a fully covered Recipe to report an empty missing list, not null")
	}

	// 120 Recipes of one Ingredient each tie on Coverage and on Missing
	// Ingredient count, so the Recipe name settles which 50 survive the cap.
	// Without a tie-break that reaches every one of them, two identical
	// requests return two different halves of the corpus.
	if last := matches[len(matches)-1].ID; last != "recipe-049" {
		t.Fatalf("cap kept up to %q, want the first fifty by Recipe name", last)
	}
}

// The same pantry over the same corpus answers with the same list, however the
// candidates happen to come back from the database. Narrowing does not order
// them by anything the ranking cares about, so this is the property that keeps
// a client's merge of the server's list with its own coherent.
func TestRecipeHandlerMatchRecipesIsDeterministicAcrossRequests(t *testing.T) {
	corpus := []models.MatchCandidate{
		{Recipe: models.RecipeSummary{ID: "pancakes", Name: "Pancakes"}, Ingredients: []string{"flour", "eggs", "milk", "butter", "sugar", "salt"}},
		{Recipe: models.RecipeSummary{ID: "crepes", Name: "Crepes"}, Ingredients: []string{"eggs", "flour", "milk", "butter"}},
		{Recipe: models.RecipeSummary{ID: "omelette-b", Name: "Omelette"}, Ingredients: []string{"eggs", "butter"}},
		{Recipe: models.RecipeSummary{ID: "omelette-a", Name: "Omelette"}, Ingredients: []string{"eggs", "butter"}},
		{Recipe: models.RecipeSummary{ID: "butter-cookies", Name: "Butter Cookies"}, Ingredients: []string{"butter", "sugar", "flour", "egg"}},
	}

	call := 0
	h := NewRecipeHandler(fakeRecipeRepo{
		getMatchCandidatesFunc: func(ctx context.Context, tokens []string, limit int) ([]models.MatchCandidate, error) {
			// A different arrival order every call, which is all narrowing
			// promises.
			rotated := append([]models.MatchCandidate{}, corpus[call%len(corpus):]...)
			rotated = append(rotated, corpus[:call%len(corpus)]...)
			call++
			return rotated, nil
		},
	})

	first := matchRequest(t, h, `["eggs", "butter", "flour", "sugar", "milk"]`).Body.String()
	for run := 1; run < len(corpus); run++ {
		if got := matchRequest(t, h, `["eggs", "butter", "flour", "sugar", "milk"]`).Body.String(); got != first {
			t.Fatalf("request %d answered\n%s\nwant\n%s", run, got, first)
		}
	}
}

func TestRecipeHandlerMatchRecipesNeedsNoAuthentication(t *testing.T) {
	h := NewRecipeHandler(fakeRecipeRepo{
		getMatchCandidatesFunc: func(ctx context.Context, tokens []string, limit int) ([]models.MatchCandidate, error) {
			return nil, nil
		},
	})

	rec := matchRequest(t, h, `["onion"]`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d without an Authorization header, got %d", http.StatusOK, rec.Code)
	}
}
