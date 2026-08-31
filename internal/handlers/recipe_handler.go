package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/thomaslgrega/bitelyapi/internal/matching"
	"github.com/thomaslgrega/bitelyapi/internal/middleware"
	"github.com/thomaslgrega/bitelyapi/internal/models"
)

type recipeRepository interface {
	GetRecipeById(ctx context.Context, id string) (models.Recipe, error)
	GetRecipesByCategory(ctx context.Context, category string) ([]models.RecipeSummary, error)
	GetRecipeFeed(ctx context.Context, limit int) ([]models.RecipeSummary, error)
	SearchRecipesByName(ctx context.Context, query string, category string, limit int) ([]models.RecipeSummary, error)
	GetRecipesByUserID(ctx context.Context, userID string) ([]models.Recipe, error)
	CreateRecipe(ctx context.Context, userID string, recipeID string, input models.CreateRecipeInput) (*models.Recipe, error)
	DeleteRecipe(ctx context.Context, id string, userID string) error
	UpdateRecipe(ctx context.Context, recipe models.Recipe, userID string) error
	GetMatchCandidates(ctx context.Context, tokens []string, limit int) ([]models.MatchCandidate, error)
}
type RecipeHandler struct {
	repo   recipeRepository
	images imageStore
}

func NewRecipeHandler(repo recipeRepository, images imageStore) *RecipeHandler {
	return &RecipeHandler{repo: repo, images: images}
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

// GetRecipes browses the corpus three ways. A Name Query matches the Recipe's
// name only — never its Ingredients, which is what pantry matching is for —
// and matches it fuzzily, so a misremembered spelling still reaches the Recipe
// (ADR-0004). A category narrows to one. Neither given, the answer is the Feed
// (ADR-0005).
//
// A Name Query and a category compose: together they search within that
// category.
func (h *RecipeHandler) GetRecipes(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	// A query of nothing but whitespace names no Recipe, so it is dropped here
	// and the request browses on whatever is left. Left in, it would reach the
	// database only to come back empty, and a request carrying a category as
	// well would answer nothing instead of that category.
	query := strings.TrimSpace(r.URL.Query().Get("name"))

	limit, err := requestedLimit(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var recipes []models.RecipeSummary
	switch {
	case query != "":
		recipes, err = h.repo.SearchRecipesByName(r.Context(), query, category, cappedAt(limit, maxNameQueryResults))
	case category != "":
		// A category listing is the one browse that answers everything it
		// finds, so it has no cap for a limit to lower. Honouring one here
		// would make the same parameter mean something different per branch.
		if limit != noLimit {
			http.Error(w, "limit does not apply when listing a category", http.StatusBadRequest)
			return
		}
		recipes, err = h.repo.GetRecipesByCategory(r.Context(), category)
	default:
		recipes, err = h.repo.GetRecipeFeed(r.Context(), cappedAt(limit, maxFeedResults))
	}
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

	// The Recipe id is minted here rather than by the database because the
	// promoted key is derived from it, and the object moves before the row is
	// written: a failure then leaves an object the lifecycle rule eats, rather
	// than a row pointing at nothing (ADR-0006).
	recipeID := uuid.NewString()

	stagedKey := input.ImageKey
	if stagedKey != "" {
		promoted, err := h.promoteImage(r.Context(), stagedKey, recipeID)
		if err != nil {
			writeImageError(w, err)
			return
		}
		input.ImageKey = promoted
	}

	recipe, err := h.repo.CreateRecipe(r.Context(), userID, recipeID, input)
	if err != nil {
		http.Error(w, "failed to create recipe", http.StatusInternalServerError)
		return
	}

	if stagedKey != "" {
		h.discardImage(r.Context(), stagedKey)
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

	h.discardImage(r.Context(), models.PromotedImageKey(id))

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

	stagedKey := recipe.ImageKey
	if stagedKey != "" {
		promoted, err := h.promoteImage(r.Context(), stagedKey, recipe.ID)
		if err != nil {
			writeImageError(w, err)
			return
		}
		recipe.ImageKey = promoted
	}

	err = h.repo.UpdateRecipe(r.Context(), recipe, userID)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "recipe not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "failed to update recipe", http.StatusInternalServerError)
		return
	}

	// A promotion overwrites the derived key, so the only object left behind
	// is the staged one. An update that keeps no image supersedes the derived
	// key itself. Either way the row has already committed.
	if stagedKey != "" {
		h.discardImage(r.Context(), stagedKey)
	} else {
		h.discardImage(r.Context(), models.PromotedImageKey(recipe.ID))
	}

	w.WriteHeader(http.StatusNoContent)
}

const (
	// maxMatches caps the response. There is no pagination: past fifty
	// Recipes the Coverage is low enough that nobody scrolls.
	maxMatches = 50

	// maxNameQueryResults caps a Name Query's answer. Like maxMatches there is
	// no pagination: a Name Query is how someone reaches a Recipe they already
	// have in mind, and past fifty the trigram tail is noise nobody scrolls.
	// These are results, not Matches — a Match answers a pantry.
	maxNameQueryResults = 50

	// candidateCeiling caps how much of the corpus narrowing may pull into
	// memory, so a pantry of common foods cannot drag the whole table in.
	candidateCeiling = 500

	// maxFeedResults caps the Feed at the same fifty the other browse
	// endpoints stop at, for the reason ADR-0005 records.
	maxFeedResults = 50

	// noLimit is what a request that asks for no cap of its own reads as.
	noLimit = 0
)

// requestedLimit reads the cap a request asks for, or noLimit when it names
// none. Zero, a fraction and an empty limit are client bugs worth reporting
// rather than quietly rounding into a page size nobody asked for.
func requestedLimit(r *http.Request) (int, error) {
	if !r.URL.Query().Has("limit") {
		return noLimit, nil
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit < 1 {
		return noLimit, errors.New("limit must be a positive whole number")
	}

	return limit, nil
}

// cappedAt lowers a browse's cap to what the request asked for. A limit past
// the cap asks for more than there is rather than for something forbidden, so
// it clamps instead of failing.
func cappedAt(limit int, ceiling int) int {
	if limit == noLimit {
		return ceiling
	}

	return min(limit, ceiling)
}

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
