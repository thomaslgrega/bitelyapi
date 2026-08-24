package models

// A MatchCandidate is a Recipe the corpus narrowed to, carried with the
// Ingredient names the matcher scores it on. Narrowing is all Postgres does:
// candidates arrive unscored and unordered.
type MatchCandidate struct {
	Recipe      RecipeSummary
	Ingredients []string
}

// A RecipeMatch is one entry of a match response: what a recipe card needs to
// render, plus which Ingredients the pantry covers and which it misses. It is
// keyed by the corpus Recipe id, so a client can deduplicate a Saved Recipe
// against its corpus original.
type RecipeMatch struct {
	RecipeSummary
	MatchedIngredients []string `json:"matched_ingredients"`
	MissingIngredients []string `json:"missing_ingredients"`
	Coverage           float64  `json:"coverage"`
}
