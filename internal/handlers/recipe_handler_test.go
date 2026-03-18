package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thomaslgrega/bitelyapi/internal/models"
)

type fakeRecipeRepo struct {
	getRecipeByIDFunc         func(ctx context.Context, id string) (models.Recipe, error)
	getRecipesByCategoryFunc  func(ctx context.Context, category string) ([]models.RecipeSummary, error)
	getRecipesByUserIDFunc    func(ctx context.Context, userID string) ([]models.Recipe, error)
	createRecipeFunc          func(ctx context.Context, userID string, input models.CreateRecipeInput) (*models.Recipe, error)
	deleteRecipeFunc          func(ctx context.Context, id string, userID string) error
	updateRecipeFunc          func(ctx context.Context, recipe models.Recipe, userID string) error
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
	t.Run("requires category", func(t *testing.T) {
		h := NewRecipeHandler(fakeRecipeRepo{})
		req := httptest.NewRequest(http.MethodGet, "/recipes", nil)
		rec := httptest.NewRecorder()

		h.GetRecipes(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("returns repository error", func(t *testing.T) {
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
