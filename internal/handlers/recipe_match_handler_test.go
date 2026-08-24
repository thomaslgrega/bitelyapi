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

func TestRecipeHandlerMatchRecipesRejectsAnEmptyPantry(t *testing.T) {
	tests := map[string]string{
		"invalid json": `{`,
		"empty list":   `[]`,
		"all blank":    `["", "   "]`,
		"not a list":   `{"pantry": ["onion"]}`,
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

func TestRecipeHandlerMatchRecipesWithAPantryThatNormalizesToNothing(t *testing.T) {
	h := NewRecipeHandler(fakeRecipeRepo{
		getMatchCandidatesFunc: func(ctx context.Context, tokens []string, limit int) ([]models.MatchCandidate, error) {
			t.Fatal("expected no repository call for a pantry with no Ingredient Terms")
			return nil, nil
		},
	})

	rec := matchRequest(t, h, `["to taste", "freshly chopped"]`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if matches := decodeMatches(t, rec); len(matches) != 0 {
		t.Fatalf("expected an empty list, got %v", matches)
	}
	if body := rec.Body.String(); body != "[]\n" {
		t.Fatalf("expected an empty JSON list, got %q", body)
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
			// that then covers nothing.
			return []models.MatchCandidate{
				{
					Recipe:      models.RecipeSummary{ID: "chowder", Name: "Corn Chowder"},
					Ingredients: []string{"cornstarch", "cream"},
				},
			}, nil
		},
	})

	rec := matchRequest(t, h, `["corn"]`)

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
