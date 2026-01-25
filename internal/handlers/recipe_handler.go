package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/thomaslgrega/bitelyapi/internal/middleware"
	"github.com/thomaslgrega/bitelyapi/internal/models"
	"github.com/thomaslgrega/bitelyapi/internal/repository"
)

type RecipeHandler struct {
	repo *repository.RecipeRepository
}

func NewRecipeHandler(repo *repository.RecipeRepository) *RecipeHandler {
	return &RecipeHandler{repo: repo}
}

func (h *RecipeHandler) GetRecipeById(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	recipe, err := h.repo.GetRecipeById(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "recipe not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to fetch recipe", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(recipe); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *RecipeHandler) GetRecipes(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	if category == "" {
		http.Error(w, "category required", http.StatusBadRequest)
		return
	}

	recipes, err := h.repo.GetRecipesByCategory(r.Context(), category)
	if err != nil {
		http.Error(w, "failed to fetch recipes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(recipes); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *RecipeHandler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var input models.CreateRecipeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if input.Name == "" || input.Category == "" {
		http.Error(w, "Name and Category are required", http.StatusBadRequest)
		return
	}

	recipe, err := h.repo.CreateRecipe(r.Context(), userID, input)
	if err != nil {
		http.Error(w, "failed to create recipe", http.StatusInternalServerError)
		return
	}

	w.WriteHeader((http.StatusCreated))
	if err := json.NewEncoder(w).Encode(recipe); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *RecipeHandler) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	err := h.repo.DeleteRecipe(r.Context(),id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "recipes not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "recipe delete failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RecipeHandler) UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	var recipe models.Recipe
	if err := json.NewDecoder(r.Body).Decode(&recipe); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	err := h.repo.UpdateRecipe(r.Context(), recipe)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "recipe not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "failed to update recipe", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}