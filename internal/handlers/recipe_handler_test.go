package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/thomaslgrega/bitelyapi/internal/models"
)

type fakeRecipeRepo struct {
	getRecipeByIDFunc        func(ctx context.Context, id string) (models.Recipe, error)
	getRecipesByCategoryFunc func(ctx context.Context, category string) ([]models.RecipeSummary, error)
	getRecipesByUserIDFunc   func(ctx context.Context, userID string) ([]models.Recipe, error)
	createRecipeFunc         func(ctx context.Context, userID string, input models.CreateRecipeInput) (*models.Recipe, error)
	deleteRecipeFunc         func(ctx context.Context, id string, userID string) error
	updateRecipeFunc         func(ctx context.Context, recipe models.Recipe, userID string) error
	getMatchCandidatesFunc   func(ctx context.Context, tokens []string, limit int) ([]models.MatchCandidate, error)
	searchRecipesByNameFunc  func(ctx context.Context, query string, category string, limit int) ([]models.RecipeSummary, error)
	getRecipeFeedFunc        func(ctx context.Context, limit int) ([]models.RecipeSummary, error)
}

func (f fakeRecipeRepo) GetRecipeById(ctx context.Context, id string) (models.Recipe, error) {
	return f.getRecipeByIDFunc(ctx, id)
}

func (f fakeRecipeRepo) GetRecipesByCategory(ctx context.Context, category string) ([]models.RecipeSummary, error) {
	return f.getRecipesByCategoryFunc(ctx, category)
}

func (f fakeRecipeRepo) GetRecipesByUserID(ctx context.Context, userID string) ([]models.Recipe, error) {
	return f.getRecipesByUserIDFunc(ctx, userID)
}

func (f fakeRecipeRepo) CreateRecipe(ctx context.Context, userID string, input models.CreateRecipeInput) (*models.Recipe, error) {
	return f.createRecipeFunc(ctx, userID, input)
}

func (f fakeRecipeRepo) DeleteRecipe(ctx context.Context, id string, userID string) error {
	return f.deleteRecipeFunc(ctx, id, userID)
}

func (f fakeRecipeRepo) UpdateRecipe(ctx context.Context, recipe models.Recipe, userID string) error {
	return f.updateRecipeFunc(ctx, recipe, userID)
}

func (f fakeRecipeRepo) SearchRecipesByName(ctx context.Context, query string, category string, limit int) ([]models.RecipeSummary, error) {
	return f.searchRecipesByNameFunc(ctx, query, category, limit)
}

func (f fakeRecipeRepo) GetRecipeFeed(ctx context.Context, limit int) ([]models.RecipeSummary, error) {
	return f.getRecipeFeedFunc(ctx, limit)
}

func (f fakeRecipeRepo) GetMatchCandidates(ctx context.Context, tokens []string, limit int) ([]models.MatchCandidate, error) {
	return f.getMatchCandidatesFunc(ctx, tokens, limit)
}

func TestRecipeHandlerGetRecipeByID(t *testing.T) {
	t.Run("requires id", func(t *testing.T) {
		h := NewRecipeHandler(fakeRecipeRepo{})
		req := httptest.NewRequest(http.MethodGet, "/recipes/", nil)
		rec := httptest.NewRecorder()

		h.GetRecipeById(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("returns not found", func(t *testing.T) {
		h := NewRecipeHandler(fakeRecipeRepo{
			getRecipeByIDFunc: func(ctx context.Context, id string) (models.Recipe, error) {
				return models.Recipe{}, sql.ErrNoRows
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/recipes/recipe-1", nil)
		req.SetPathValue("id", "recipe-1")
		rec := httptest.NewRecorder()

		h.GetRecipeById(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})

	t.Run("returns recipe json", func(t *testing.T) {
		h := NewRecipeHandler(fakeRecipeRepo{
			getRecipeByIDFunc: func(ctx context.Context, id string) (models.Recipe, error) {
				return models.Recipe{ID: id, Name: "Soup", Category: "dinner"}, nil
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/recipes/recipe-1", nil)
		req.SetPathValue("id", "recipe-1")
		rec := httptest.NewRecorder()

		h.GetRecipeById(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		var recipe models.Recipe
		if err := json.Unmarshal(rec.Body.Bytes(), &recipe); err != nil {
			t.Fatalf("failed to decode recipe: %v", err)
		}
		if recipe.ID != "recipe-1" {
			t.Fatalf("expected path id to be returned, got %q", recipe.ID)
		}
	})
}

func TestRecipeHandlerGetRecipes(t *testing.T) {
	t.Run("answers neither a category nor a Name Query with the Feed", func(t *testing.T) {
		h := NewRecipeHandler(fakeRecipeRepo{
			getRecipeFeedFunc: func(ctx context.Context, limit int) ([]models.RecipeSummary, error) {
				if limit != maxFeedResults {
					t.Fatalf("expected limit %d, got %d", maxFeedResults, limit)
				}
				return []models.RecipeSummary{{ID: "recipe-1", Name: "Green Shakshuka", Category: "breakfast"}}, nil
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/recipes", nil)
		rec := httptest.NewRecorder()

		h.GetRecipes(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		var recipes []models.RecipeSummary
		if err := json.Unmarshal(rec.Body.Bytes(), &recipes); err != nil {
			t.Fatalf("failed to decode recipes: %v", err)
		}
		if len(recipes) != 1 || recipes[0].Name != "Green Shakshuka" {
			t.Fatalf("unexpected recipes: %#v", recipes)
		}
	})

	t.Run("answers a blank Name Query with the Feed", func(t *testing.T) {
		called := false
		h := NewRecipeHandler(fakeRecipeRepo{
			getRecipeFeedFunc: func(ctx context.Context, limit int) ([]models.RecipeSummary, error) {
				called = true
				return []models.RecipeSummary{}, nil
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/recipes?name=%20%20", nil)
		rec := httptest.NewRecorder()

		h.GetRecipes(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if !called {
			t.Fatal("expected a blank Name Query to fall through to the Feed")
		}
	})

	t.Run("lets a limit shorten the Feed", func(t *testing.T) {
		h := NewRecipeHandler(fakeRecipeRepo{
			getRecipeFeedFunc: func(ctx context.Context, limit int) ([]models.RecipeSummary, error) {
				if limit != 5 {
					t.Fatalf("expected limit 5, got %d", limit)
				}
				return []models.RecipeSummary{}, nil
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/recipes?limit=5", nil)
		rec := httptest.NewRecorder()

		h.GetRecipes(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("clamps a limit past the cap", func(t *testing.T) {
		h := NewRecipeHandler(fakeRecipeRepo{
			getRecipeFeedFunc: func(ctx context.Context, limit int) ([]models.RecipeSummary, error) {
				if limit != maxFeedResults {
					t.Fatalf("expected limit %d, got %d", maxFeedResults, limit)
				}
				return []models.RecipeSummary{}, nil
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/recipes?limit=500", nil)
		rec := httptest.NewRecorder()

		h.GetRecipes(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("rejects a limit that is not a positive whole number", func(t *testing.T) {
		for _, limit := range []string{"abc", "0", "-3", "2.5", ""} {
			h := NewRecipeHandler(fakeRecipeRepo{})
			req := httptest.NewRequest(http.MethodGet, "/recipes?limit="+limit, nil)
			rec := httptest.NewRecorder()

			h.GetRecipes(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("limit %q: expected status %d, got %d", limit, http.StatusBadRequest, rec.Code)
			}
		}
	})

	t.Run("lets a limit shorten a Name Query too", func(t *testing.T) {
		h := NewRecipeHandler(fakeRecipeRepo{
			searchRecipesByNameFunc: func(ctx context.Context, query string, category string, limit int) ([]models.RecipeSummary, error) {
				if limit != 5 {
					t.Fatalf("expected limit 5, got %d", limit)
				}
				return []models.RecipeSummary{}, nil
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/recipes?name=shakshuka&limit=5", nil)
		rec := httptest.NewRecorder()

		h.GetRecipes(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("refuses a limit on a category listing, which has no cap to lower", func(t *testing.T) {
		h := NewRecipeHandler(fakeRecipeRepo{})
		req := httptest.NewRequest(http.MethodGet, "/recipes?category=dinner&limit=5", nil)
		rec := httptest.NewRecorder()

		h.GetRecipes(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("returns a Feed repository error", func(t *testing.T) {
		h := NewRecipeHandler(fakeRecipeRepo{
			getRecipeFeedFunc: func(ctx context.Context, limit int) ([]models.RecipeSummary, error) {
				return nil, errors.New("db down")
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/recipes", nil)
		rec := httptest.NewRecorder()

		h.GetRecipes(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})

	t.Run("lists a category rather than the Feed", func(t *testing.T) {
		h := NewRecipeHandler(fakeRecipeRepo{
			getRecipesByCategoryFunc: func(ctx context.Context, category string) ([]models.RecipeSummary, error) {
				if category != "dinner" {
					t.Fatalf("expected category dinner, got %q", category)
				}
				return []models.RecipeSummary{{ID: "recipe-1", Name: "Soup", Category: "dinner"}}, nil
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/recipes?category=dinner", nil)
		rec := httptest.NewRecorder()

		h.GetRecipes(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		var recipes []models.RecipeSummary
		if err := json.Unmarshal(rec.Body.Bytes(), &recipes); err != nil {
			t.Fatalf("failed to decode recipes: %v", err)
		}
		if len(recipes) != 1 || recipes[0].Name != "Soup" {
			t.Fatalf("unexpected recipes: %#v", recipes)
		}
	})

	t.Run("returns a category repository error", func(t *testing.T) {
		h := NewRecipeHandler(fakeRecipeRepo{
			getRecipesByCategoryFunc: func(ctx context.Context, category string) ([]models.RecipeSummary, error) {
				return nil, errors.New("db down")
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/recipes?category=dinner", nil)
		rec := httptest.NewRecorder()

		h.GetRecipes(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})

	t.Run("searches by name when a query is given", func(t *testing.T) {
		h := NewRecipeHandler(fakeRecipeRepo{
			searchRecipesByNameFunc: func(ctx context.Context, query string, category string, limit int) ([]models.RecipeSummary, error) {
				if query != "shakshuka" {
					t.Fatalf("expected trimmed query, got %q", query)
				}
				if category != "" {
					t.Fatalf("expected no category, got %q", category)
				}
				if limit != maxNameQueryResults {
					t.Fatalf("expected limit %d, got %d", maxNameQueryResults, limit)
				}
				return []models.RecipeSummary{{ID: "recipe-1", Name: "Green Shakshuka", Category: "breakfast"}}, nil
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/recipes?name=%20shakshuka%20", nil)
		rec := httptest.NewRecorder()

		h.GetRecipes(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		var recipes []models.RecipeSummary
		if err := json.Unmarshal(rec.Body.Bytes(), &recipes); err != nil {
			t.Fatalf("failed to decode recipes: %v", err)
		}
		if len(recipes) != 1 || recipes[0].Name != "Green Shakshuka" {
			t.Fatalf("unexpected recipes: %#v", recipes)
		}
	})

	t.Run("composes a Name Query with a category", func(t *testing.T) {
		h := NewRecipeHandler(fakeRecipeRepo{
			searchRecipesByNameFunc: func(ctx context.Context, query string, category string, limit int) ([]models.RecipeSummary, error) {
				if query != "shakshuka" || category != "breakfast" {
					t.Fatalf("unexpected search: query %q, category %q", query, category)
				}
				return []models.RecipeSummary{}, nil
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/recipes?name=shakshuka&category=breakfast", nil)
		rec := httptest.NewRecorder()

		h.GetRecipes(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if body := rec.Body.String(); body != "[]\n" {
			t.Fatalf("expected an empty list, got %q", body)
		}
	})

	t.Run("returns a name search repository error", func(t *testing.T) {
		h := NewRecipeHandler(fakeRecipeRepo{
			searchRecipesByNameFunc: func(ctx context.Context, query string, category string, limit int) ([]models.RecipeSummary, error) {
				return nil, errors.New("db down")
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/recipes?name=shakshuka", nil)
		rec := httptest.NewRecorder()

		h.GetRecipes(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})
}

func TestRecipeHandlerGetMyRecipes(t *testing.T) {
	t.Run("requires auth context", func(t *testing.T) {
		h := NewRecipeHandler(fakeRecipeRepo{})
		req := httptest.NewRequest(http.MethodGet, "/me/recipes", nil)
		rec := httptest.NewRecorder()

		h.GetMyRecipes(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("returns recipes for authenticated user", func(t *testing.T) {
		h := NewRecipeHandler(fakeRecipeRepo{
			getRecipesByUserIDFunc: func(ctx context.Context, userID string) ([]models.Recipe, error) {
				if userID != "user-1" {
					t.Fatalf("expected user-1, got %q", userID)
				}
				return []models.Recipe{{ID: "recipe-1", Name: "Soup", Category: "dinner"}}, nil
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/me/recipes", nil)
		rec := authedRequest(t, req, h.GetMyRecipes)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})
}

func TestRecipeHandlerCreateRecipe(t *testing.T) {
	t.Run("requires auth context", func(t *testing.T) {
		h := NewRecipeHandler(fakeRecipeRepo{})
		req := httptest.NewRequest(http.MethodPost, "/recipes", bytes.NewBufferString(`{}`))
		rec := httptest.NewRecorder()

		h.CreateRecipe(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("rejects invalid json", func(t *testing.T) {
		h := NewRecipeHandler(fakeRecipeRepo{})
		req := httptest.NewRequest(http.MethodPost, "/recipes", bytes.NewBufferString("{"))
		rec := authedRequest(t, req, h.CreateRecipe)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("requires name and category", func(t *testing.T) {
		h := NewRecipeHandler(fakeRecipeRepo{})
		req := httptest.NewRequest(http.MethodPost, "/recipes", bytes.NewBufferString(`{"name":"Soup"}`))
		rec := authedRequest(t, req, h.CreateRecipe)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("creates recipe", func(t *testing.T) {
		h := NewRecipeHandler(fakeRecipeRepo{
			createRecipeFunc: func(ctx context.Context, userID string, input models.CreateRecipeInput) (*models.Recipe, error) {
				if userID != "user-1" {
					t.Fatalf("expected user-1, got %q", userID)
				}
				if input.Name != "Soup" || input.Category != "dinner" {
					t.Fatalf("unexpected input: %#v", input)
				}
				return &models.Recipe{ID: "recipe-1", UserID: userID, Name: input.Name, Category: input.Category}, nil
			},
		})
		req := httptest.NewRequest(http.MethodPost, "/recipes", bytes.NewBufferString(`{"name":"Soup","category":"dinner"}`))
		rec := authedRequest(t, req, h.CreateRecipe)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
		}
	})
}

func TestRecipeHandlerDeleteRecipe(t *testing.T) {
	t.Run("maps missing rows to not found", func(t *testing.T) {
		h := NewRecipeHandler(fakeRecipeRepo{
			deleteRecipeFunc: func(ctx context.Context, id string, userID string) error {
				return sql.ErrNoRows
			},
		})
		req := httptest.NewRequest(http.MethodDelete, "/recipes/recipe-1", nil)
		req.SetPathValue("id", "recipe-1")
		rec := authedRequest(t, req, h.DeleteRecipe)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})

	t.Run("returns no content on success", func(t *testing.T) {
		h := NewRecipeHandler(fakeRecipeRepo{
			deleteRecipeFunc: func(ctx context.Context, id string, userID string) error {
				if id != "recipe-1" || userID != "user-1" {
					t.Fatalf("unexpected delete args id=%q userID=%q", id, userID)
				}
				return nil
			},
		})
		req := httptest.NewRequest(http.MethodDelete, "/recipes/recipe-1", nil)
		req.SetPathValue("id", "recipe-1")
		rec := authedRequest(t, req, h.DeleteRecipe)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
		}
	})
}

func TestRecipeHandlerUpdateRecipe(t *testing.T) {
	t.Run("rejects invalid json", func(t *testing.T) {
		h := NewRecipeHandler(fakeRecipeRepo{})
		req := httptest.NewRequest(http.MethodPut, "/recipes/recipe-1", bytes.NewBufferString("{"))
		rec := authedRequest(t, req, h.UpdateRecipe)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("path id overrides body id", func(t *testing.T) {
		h := NewRecipeHandler(fakeRecipeRepo{
			updateRecipeFunc: func(ctx context.Context, recipe models.Recipe, userID string) error {
				if recipe.ID != "recipe-from-path" {
					t.Fatalf("expected path id to win, got %q", recipe.ID)
				}
				if userID != "user-1" {
					t.Fatalf("expected user-1, got %q", userID)
				}
				return nil
			},
		})
		req := httptest.NewRequest(http.MethodPut, "/recipes/recipe-from-path", bytes.NewBufferString(`{"id":"recipe-from-body","name":"Soup","category":"dinner"}`))
		req.SetPathValue("id", "recipe-from-path")
		rec := authedRequest(t, req, h.UpdateRecipe)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
		}
	})

	t.Run("maps missing rows to not found", func(t *testing.T) {
		h := NewRecipeHandler(fakeRecipeRepo{
			updateRecipeFunc: func(ctx context.Context, recipe models.Recipe, userID string) error {
				return sql.ErrNoRows
			},
		})
		req := httptest.NewRequest(http.MethodPut, "/recipes/recipe-1", bytes.NewBufferString(`{"name":"Soup","category":"dinner"}`))
		req.SetPathValue("id", "recipe-1")
		rec := authedRequest(t, req, h.UpdateRecipe)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})
}

// recipeStore is a fake repository that keeps what it is given, so a test can
// create a recipe and then fetch it back the way a real client would.
type recipeStore struct {
	fakeRecipeRepo
	recipes map[string]models.Recipe
}

func newRecipeStore() *recipeStore {
	return &recipeStore{recipes: make(map[string]models.Recipe)}
}

func (s *recipeStore) CreateRecipe(ctx context.Context, userID string, input models.CreateRecipeInput) (*models.Recipe, error) {
	ingredients := make([]models.Ingredient, 0, len(input.Ingredients))
	for i, ingredient := range input.Ingredients {
		ingredients = append(ingredients, models.Ingredient{
			ID:          fmt.Sprintf("ingredient-%d", i+1),
			Name:        ingredient.Name,
			Measurement: ingredient.Measurement,
		})
	}

	recipe := models.Recipe{
		ID:            "recipe-1",
		UserID:        userID,
		Name:          input.Name,
		Category:      input.Category,
		Instructions:  input.Instructions,
		ThumbnailUrl:  input.ThumbnailUrl,
		Ingredients:   ingredients,
		Calories:      input.Calories,
		TotalCookTime: input.TotalCookTime,
	}
	s.recipes[recipe.ID] = recipe
	return &recipe, nil
}

func (s *recipeStore) GetRecipeById(ctx context.Context, id string) (models.Recipe, error) {
	recipe, ok := s.recipes[id]
	if !ok {
		return models.Recipe{}, sql.ErrNoRows
	}
	return recipe, nil
}

func TestRecipeHandlerCreateRecipeTrimsNames(t *testing.T) {
	t.Run("create response carries the stored names", func(t *testing.T) {
		h := NewRecipeHandler(newRecipeStore())

		body := `{"name":"  Chicken Dinner ","category":"dinner","ingredients":[{"name":"  Chicken Breast ","measurement":"2 lbs"}]}`
		req := httptest.NewRequest(http.MethodPost, "/recipes", bytes.NewBufferString(body))
		rec := authedRequest(t, req, h.CreateRecipe)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
		}

		var created models.Recipe
		if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
			t.Fatalf("failed to decode created recipe: %v", err)
		}
		if created.Name != "Chicken Dinner" {
			t.Fatalf("expected trimmed recipe name, got %q", created.Name)
		}
		if len(created.Ingredients) != 1 {
			t.Fatalf("expected 1 ingredient, got %d", len(created.Ingredients))
		}
		if created.Ingredients[0].Name != "Chicken Breast" {
			t.Fatalf("expected trimmed ingredient name, got %q", created.Ingredients[0].Name)
		}

		fetchReq := httptest.NewRequest(http.MethodGet, "/recipes/"+created.ID, nil)
		fetchReq.SetPathValue("id", created.ID)
		fetchRec := httptest.NewRecorder()
		h.GetRecipeById(fetchRec, fetchReq)

		if fetchRec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, fetchRec.Code)
		}
		var fetched models.Recipe
		if err := json.Unmarshal(fetchRec.Body.Bytes(), &fetched); err != nil {
			t.Fatalf("failed to decode fetched recipe: %v", err)
		}
		if !reflect.DeepEqual(created, fetched) {
			t.Fatalf("create response %#v does not match fetch %#v", created, fetched)
		}
	})

	t.Run("rejects a whitespace-only name", func(t *testing.T) {
		h := NewRecipeHandler(newRecipeStore())
		req := httptest.NewRequest(http.MethodPost, "/recipes", bytes.NewBufferString(`{"name":"   ","category":"dinner"}`))
		rec := authedRequest(t, req, h.CreateRecipe)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})
}

func TestRecipeHandlerUpdateRecipeTrimsNames(t *testing.T) {
	t.Run("trims recipe and ingredient names before storing", func(t *testing.T) {
		h := NewRecipeHandler(fakeRecipeRepo{
			updateRecipeFunc: func(ctx context.Context, recipe models.Recipe, userID string) error {
				if recipe.Name != "Chicken Dinner" {
					t.Fatalf("expected trimmed recipe name, got %q", recipe.Name)
				}
				if len(recipe.Ingredients) != 1 || recipe.Ingredients[0].Name != "Chicken Breast" {
					t.Fatalf("expected trimmed ingredient name, got %#v", recipe.Ingredients)
				}
				return nil
			},
		})

		body := `{"name":"  Chicken Dinner ","category":"dinner","ingredients":[{"id":"ingredient-1","name":"  Chicken Breast ","measurement":"2 lbs"}]}`
		req := httptest.NewRequest(http.MethodPut, "/recipes/recipe-1", bytes.NewBufferString(body))
		req.SetPathValue("id", "recipe-1")
		rec := authedRequest(t, req, h.UpdateRecipe)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
		}
	})
}
