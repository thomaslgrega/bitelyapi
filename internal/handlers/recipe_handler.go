package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/thomaslgrega/bitelyapi/internal/models"
	"github.com/thomaslgrega/bitelyapi/internal/repository"
)

type RecipeHandler struct {
	repo *repository.RecipeRepository
}

func NewRecipeHandler(repo *repository.RecipeRepository) *RecipeHandler {
	return &RecipeHandler{repo: repo}
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
	json.NewEncoder(w).Encode(recipes)
}

func (h *RecipeHandler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	var input models.CreateRecipeInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if input.Name == "" || input.Category == "" {
		http.Error(w, "Name and Category are required", http.StatusBadRequest)
		return
	}

	recipe, err := h.repo.CreateRecipe(r.Context(), input)
	if err != nil {
		log.Println(err)
		http.Error(w, "failed to create recipe", http.StatusInternalServerError)
		return
	}

	w.WriteHeader((http.StatusCreated))
	json.NewEncoder(w).Encode(recipe)
}

// func UpdateRecipe(w http.ResponseWriter, r *http.Request) {

// }

// func DeleteRecipe(w http.ResponseWriter, r *http.Request) {

// }