package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/thomaslgrega/bitelyapi/internal/matching"
	"github.com/thomaslgrega/bitelyapi/internal/middleware"
	"github.com/thomaslgrega/bitelyapi/internal/models"
)

type recipeRepository interface {
	GetRecipeById(ctx context.Context, id string) (models.Recipe, error)
	GetRecipesByCategory(ctx context.Context, category string) ([]models.RecipeSummary, error)
	GetRecipesByUserID(ctx context.Context, userID string) ([]models.Recipe, error)
	CreateRecipe(ctx context.Context, userID string, input models.CreateRecipeInput) (*models.Recipe, error)
	DeleteRecipe(ctx context.Context, id string, userID string) error
	UpdateRecipe(ctx context.Context, recipe models.Recipe, userID string) error
	GetMatchCandidates(ctx context.Context, tokens []string, limit int) ([]models.MatchCandidate, error)
}
type RecipeHandler struct {
	repo recipeRepository
}

func NewRecipeHandler(repo recipeRepository) *RecipeHandler {
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

func (h *RecipeHandler) GetMyRecipes(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	recipes, err := h.repo.GetRecipesByUserID(r.Context(), userID)
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
	input.TrimNames()

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
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	err = h.repo.DeleteRecipe(r.Context(), id, userID)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "recipe not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "recipe delete failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RecipeHandler) UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var recipe models.Recipe
	if err := json.NewDecoder(r.Body).Decode(&recipe); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	recipe.TrimNames()

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}
	recipe.ID = id

	err = h.repo.UpdateRecipe(r.Context(), recipe, userID)
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

const (
	// maxMatches caps the response. There is no pagination: past fifty
	// Recipes the Coverage is low enough that nobody scrolls.
	maxMatches = 50

	// candidateCeiling caps how much of the corpus narrowing may pull into
	// memory, so a pantry of common foods cannot drag the whole table in.
	candidateCeiling = 500
)

// MatchRecipes answers a list of raw, un-normalized Pantry Items with the
// Recipes the user could cook, best fit first. Postgres narrows the corpus;
// the matching package does all the scoring and ordering.
func (h *RecipeHandler) MatchRecipes(w http.ResponseWriter, r *http.Request) {
	var pantryItems []string
	if err := json.NewDecoder(r.Body).Decode(&pantryItems); err != nil {
		http.Error(w, "invalid JSON: expected a list of ingredient strings", http.StatusBadRequest)
		return
	}

	// Blanks are stripped before the list is measured, not after, so a list of
	// nothing but blank strings is rejected the same as an empty one. ADR-0003
	// records why: such a request carries no Pantry Item at all. Moving this
	// below the check turns those requests into an empty 200.
	submitted := make([]string, 0, len(pantryItems))
	for _, item := range pantryItems {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			submitted = append(submitted, trimmed)
		}
	}
	if len(submitted) == 0 {
		http.Error(w, "at least one ingredient is required", http.StatusBadRequest)
		return
	}

	// A pantry of nothing but measurements and descriptors has no Ingredient
	// Terms to narrow on, and so no Matches. That is an empty answer, not an
	// error, and it does not need the database to say so.
	tokens := matching.PantryTokens(submitted)
	if len(tokens) == 0 {
		writeMatches(w, []models.RecipeMatch{})
		return
	}

	candidates, err := h.repo.GetMatchCandidates(r.Context(), tokens, candidateCeiling)
	if err != nil {
		http.Error(w, "failed to match recipes", http.StatusInternalServerError)
		return
	}

	matchable := make([]matching.Candidate, 0, len(candidates))
	summaries := make(map[string]models.RecipeSummary, len(candidates))
	for _, candidate := range candidates {
		matchable = append(matchable, matching.Candidate{
			RecipeID:        candidate.Recipe.ID,
			Name:            candidate.Recipe.Name,
			IngredientNames: candidate.Ingredients,
		})
		summaries[candidate.Recipe.ID] = candidate.Recipe
	}

	ranked := matching.Rank(submitted, matchable)
	if len(ranked) > maxMatches {
		ranked = ranked[:maxMatches]
	}

	matches := make([]models.RecipeMatch, 0, len(ranked))
	for _, match := range ranked {
		missing := match.Missing
		if missing == nil {
			// A fully covered Recipe misses nothing; say so with an empty list
			// rather than a null a client has to guard.
			missing = []string{}
		}

		matches = append(matches, models.RecipeMatch{
			RecipeSummary:      summaries[match.RecipeID],
			MatchedIngredients: match.Matched,
			MissingIngredients: missing,
			Coverage:           match.Coverage(),
		})
	}

	writeMatches(w, matches)
}

func writeMatches(w http.ResponseWriter, matches []models.RecipeMatch) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(matches); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
