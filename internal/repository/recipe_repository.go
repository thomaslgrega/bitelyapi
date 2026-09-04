package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"github.com/thomaslgrega/bitelyapi/internal/models"
)

type RecipeRepository struct {
	db *sql.DB

	// images composes the public URL for a Recipe Image. It lives here because
	// scanning is the one boundary every row crosses on its way to a model, so
	// no read path can answer a stored key by mistake (ADR-0006).
	images models.ImageLocator
}

func NewRecipeRepository(db *sql.DB, images models.ImageLocator) *RecipeRepository {
	return &RecipeRepository{db: db, images: images}
}

// GetRecipeAuthor answers who authored a Recipe. An update reads it before
// promoting an image, because the promoted key derives from the Recipe id and
// a stranger must not reach it at all (ADR-0006).
func (r *RecipeRepository) GetRecipeAuthor(ctx context.Context, id string) (string, error) {
	var author string
	if err := r.db.QueryRowContext(ctx, "SELECT user_id FROM recipes WHERE id = $1", id).Scan(&author); err != nil {
		return "", err
	}

	return author, nil
}

func (r *RecipeRepository) GetRecipesByUserID(ctx context.Context, userID string) ([]models.Recipe, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, category, image_key, calories, total_cook_time
		FROM recipes
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}

	var recipes = make([]models.Recipe, 0)
	for rows.Next() {
		var recipe models.Recipe
		if err := rows.Scan(
			&recipe.ID,
			&recipe.Name,
			&recipe.Category,
			&recipe.ImageKey,
			&recipe.Calories,
			&recipe.TotalCookTime,
		); err != nil {
			return nil, err
		}
		recipe.ResolveImage(r.images)
		recipes = append(recipes, recipe)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return recipes, nil
}

func (r *RecipeRepository) GetRecipeById(ctx context.Context, id string) (models.Recipe, error) {
	row := r.db.QueryRowContext(
		ctx,
		"SELECT id, user_id, name, category, instructions, image_key, calories, total_cook_time FROM recipes WHERE id = $1",
		id,
	)

	var recipe models.Recipe
	if err := row.Scan(
		&recipe.ID,
		&recipe.UserID,
		&recipe.Name,
		&recipe.Category,
		&recipe.Instructions,
		&recipe.ImageKey,
		&recipe.Calories,
		&recipe.TotalCookTime,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return recipe, err
		}
		return recipe, err
	}

	rows, err := r.db.QueryContext(ctx, "SELECT id, name, measurement FROM ingredients WHERE recipe_id = $1", id)
	if err != nil {
		return recipe, err
	}

	defer rows.Close()
	for rows.Next() {
		var ingredient models.Ingredient
		if err := rows.Scan(&ingredient.ID, &ingredient.Name, &ingredient.Measurement); err != nil {
			return recipe, err
		}

		recipe.Ingredients = append(recipe.Ingredients, ingredient)
	}

	if err := rows.Err(); err != nil {
		return recipe, err
	}

	recipe.ResolveImage(r.images)

	return recipe, nil
}

// recipeSummaryColumns is the RecipeSummary projection every browse shares,
// in the order scanRecipeSummaries reads them.
const recipeSummaryColumns = "id, name, category, image_key, calories, total_cook_time"

func scanRecipeSummaries(rows *sql.Rows, images models.ImageLocator) ([]models.RecipeSummary, error) {
	recipes := make([]models.RecipeSummary, 0)
	for rows.Next() {
		var recipe models.RecipeSummary
		if err := rows.Scan(&recipe.ID, &recipe.Name, &recipe.Category, &recipe.ImageKey, &recipe.Calories, &recipe.TotalCookTime); err != nil {
			return nil, err
		}
		recipe.ResolveImage(images)
		recipes = append(recipes, recipe)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return recipes, nil
}

func (r *RecipeRepository) GetRecipesByCategory(ctx context.Context, category string) ([]models.RecipeSummary, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT "+recipeSummaryColumns+" FROM recipes WHERE category = $1",
		category,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRecipeSummaries(rows, r.images)
}

// GetRecipeFeed answers the Feed with the Shared Recipes shared most
// recently, newest first, capped at the limit (ADR-0005).
func (r *RecipeRepository) GetRecipeFeed(ctx context.Context, limit int) ([]models.RecipeSummary, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+recipeSummaryColumns+`
		FROM recipes
		ORDER BY created_at DESC NULLS LAST, id DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRecipeSummaries(rows, r.images)
}

// nameQueryThreshold is how similar a Name Query has to be to some run of
// words in a Recipe's name to reach it. It is far tighter than
// narrowingThreshold, and for the opposite reason: nothing re-scores these
// rows afterwards, so whatever clears the threshold is the answer the user
// reads. 0.3 is where names that merely share a stem start arriving — it
// answers "banana" with Chana Masala and "chicken" with Chocolate Chip
// Cookies. 0.5 drops those and still clears every misspelling worth serving:
// shakshouka against shakshuka scores 0.615, focacia against focaccia 0.700,
// ratatouile against ratatouille 0.818.
const nameQueryThreshold = 0.5

// SearchRecipesByName answers a Name Query with the Shared Recipes whose name
// it reaches, closest first. It matches on the name alone: searching
// Ingredients as well is what pantry matching already does, and folding it in
// here would answer "chicken" with Recipes that are not named for it
// (ADR-0004).
//
// Trigram word similarity carries both halves of what a search box gets.
// Misspellings are the obvious half. Half-typed queries are the less obvious
// one: word_similarity scores its query against the best-matching run of words
// in the name rather than the whole name, so "shak" reaches "Green Shakshuka"
// at 0.8 and needs no separate prefix or substring match beside it. What that
// gives up is matching inside a word — "toui" does not reach "Ratatouille" —
// which is the boundary a word-anchored search should have anyway.
//
// An empty category means every category. The limit caps the answer.
func (r *RecipeRepository) SearchRecipesByName(ctx context.Context, query string, category string, limit int) ([]models.RecipeSummary, error) {
	// The word-similarity threshold is a session setting, so the query runs in
	// a transaction that can SET LOCAL it rather than depend on the server
	// default.
	transaction, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer transaction.Rollback()

	// set_config with is_local true rather than SET LOCAL, because SET takes no
	// query parameters. It takes the value as text, hence the formatting.
	threshold := strconv.FormatFloat(nameQueryThreshold, 'f', -1, 64)
	if _, err := transaction.ExecContext(ctx, "SELECT set_config('pg_trgm.word_similarity_threshold', $1, true)", threshold); err != nil {
		return nil, err
	}

	// Normalized here the same way name_norm is generated, rather than leaning
	// on pg_trgm to fold case itself or on the caller to have trimmed. Untrimmed
	// input would otherwise score against the trigrams its own padding adds.
	normalized := strings.ToLower(strings.TrimSpace(query))
	rows, err := transaction.QueryContext(ctx, `
		SELECT `+recipeSummaryColumns+`
		FROM recipes
		WHERE ($2 = '' OR category = $2)
		  AND $1 <% name_norm
		ORDER BY word_similarity($1, name_norm) DESC, name, id
		LIMIT $3
	`, normalized, category, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recipes, err := scanRecipeSummaries(rows, r.images)
	if err != nil {
		return nil, err
	}

	if err := transaction.Commit(); err != nil {
		return nil, err
	}

	return recipes, nil
}

// CreateRecipe writes a Recipe under an id the caller minted. The id arrives
// rather than being defaulted by Postgres because a Recipe Image is promoted
// to a key derived from it before the row exists (ADR-0006).
func (r *RecipeRepository) CreateRecipe(ctx context.Context, userID string, recipeID string, input models.CreateRecipeInput) (*models.Recipe, error) {
	transaction, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer transaction.Rollback()

	_, err = transaction.ExecContext(ctx, `
		INSERT INTO recipes (id, user_id, name, category, instructions, image_key, calories, total_cook_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, recipeID, userID, input.Name, input.Category, input.Instructions, input.ImageKey, input.Calories, input.TotalCookTime)
	if err != nil {
		return nil, err
	}

	ingredients := make([]models.Ingredient, 0, len(input.Ingredients))
	for _, ingredient := range input.Ingredients {
		var ingredientID string
		err := transaction.QueryRowContext(ctx, `
			INSERT INTO ingredients (recipe_id, name, measurement)
			VALUES ($1, $2, $3)
			RETURNING id
		`, recipeID, ingredient.Name, ingredient.Measurement).Scan(&ingredientID)
		if err != nil {
			return nil, err
		}

		ingredients = append(ingredients, models.Ingredient{
			ID:          ingredientID,
			Name:        ingredient.Name,
			Measurement: ingredient.Measurement,
		})
	}

	if err := transaction.Commit(); err != nil {
		return nil, err
	}

	created := &models.Recipe{
		ID:            recipeID,
		UserID:        userID,
		Name:          input.Name,
		Category:      input.Category,
		Instructions:  input.Instructions,
		ImageKey:      input.ImageKey,
		Ingredients:   ingredients,
		Calories:      input.Calories,
		TotalCookTime: input.TotalCookTime,
	}
	created.ResolveImage(r.images)

	return created, nil
}

// DeleteRecipe removes a Recipe and answers the image key its row held. The
// key comes back from the delete itself rather than from a read before it, so
// an image written in between is the one the caller discards (ADR-0006).
func (r *RecipeRepository) DeleteRecipe(ctx context.Context, id string, userID string) (string, error) {
	var deleted string
	err := r.db.QueryRowContext(
		ctx,
		"DELETE FROM recipes WHERE id = $1 AND user_id = $2 RETURNING image_key",
		id, userID,
	).Scan(&deleted)
	if err != nil {
		return "", err
	}

	return deleted, nil
}

// UpdateRecipe writes a Recipe's text. It touches no image key: a Recipe Image
// is written through its own sub-resource, so a field this write omits can no
// longer destroy one (ADR-0006).
func (r *RecipeRepository) UpdateRecipe(ctx context.Context, input models.UpdateRecipeInput, userID string) error {
	transaction, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer transaction.Rollback()

	var updated string
	err = transaction.QueryRowContext(ctx, `
		UPDATE recipes
		SET name = $1, category = $2, instructions = $3, calories = $4, total_cook_time = $5
		WHERE id = $6 AND user_id = $7
		RETURNING id
	`, input.Name, input.Category, input.Instructions, input.Calories, input.TotalCookTime, input.ID, userID).
		Scan(&updated)
	if err != nil {
		return err
	}

	_, err = transaction.ExecContext(ctx, "DELETE FROM ingredients WHERE recipe_id = $1", input.ID)
	if err != nil {
		return err
	}

	for _, ingredient := range input.Ingredients {
		_, err := transaction.ExecContext(ctx, `
			INSERT INTO ingredients (id, recipe_id, name, measurement)
			VALUES ($1, $2, $3, $4)
		`, ingredient.ID, input.ID, ingredient.Name, ingredient.Measurement)

		if err != nil {
			return err
		}
	}

	return transaction.Commit()
}

// SetRecipeImage points a Recipe at a promoted object, and answers both the URL
// it now serves and the key its row stopped naming.
func (r *RecipeRepository) SetRecipeImage(ctx context.Context, recipeID string, userID string, key string) (string, string, error) {
	superseded, err := r.writeImageKey(ctx, recipeID, userID, key)
	if err != nil {
		return "", "", err
	}

	return r.images.URLFor(key), superseded, nil
}

// ClearRecipeImage leaves a Recipe with no image and answers the key its row
// stopped naming.
func (r *RecipeRepository) ClearRecipeImage(ctx context.Context, recipeID string, userID string) (string, error) {
	return r.writeImageKey(ctx, recipeID, userID, "")
}

// writeImageKey is the one write that moves a Recipe Image. The prior key is
// read under the write's own lock rather than by the caller beforehand, so two
// writes racing to replace the same image each discard the one they actually
// superseded (ADR-0006).
func (r *RecipeRepository) writeImageKey(ctx context.Context, recipeID string, userID string, key string) (string, error) {
	var superseded string
	err := r.db.QueryRowContext(ctx, `
		UPDATE recipes
		SET image_key = $1
		FROM (SELECT id, image_key FROM recipes WHERE id = $2 AND user_id = $3 FOR UPDATE) prior
		WHERE recipes.id = prior.id
		RETURNING prior.image_key
	`, key, recipeID, userID).Scan(&superseded)
	if err != nil {
		return "", err
	}

	return superseded, nil
}

// narrowingThreshold is how similar an Ingredient name has to be to a pantry
// token to make a Recipe a candidate. It is deliberately looser than scoring:
// narrowing costs recall when it is too tight, and the matching package
// re-scores every candidate from scratch, so being generous here is free.
const narrowingThreshold = 0.3

// GetMatchCandidates narrows the corpus to Recipes with at least one
// Ingredient whose name is trigram-similar to one of the pantry tokens, and
// returns them with their Ingredient names attached. It applies no scoring and
// no ordering — that is the matching package's job. The limit caps how many
// Recipes come back, so a pantry of common foods cannot pull the whole corpus
// into memory. Which Recipes survive that cap is arbitrary — nothing here
// ranks them — but ordering the cap keeps it stable between identical
// requests.
//
// Ingredients come back in the order they were stored. The table records no
// authored position, so that is the closest thing to the order the Author
// wrote them in.
func (r *RecipeRepository) GetMatchCandidates(ctx context.Context, tokens []string, limit int) ([]models.MatchCandidate, error) {
	// The word-similarity threshold is a session setting, so the query runs in
	// a transaction that can SET LOCAL it rather than depend on the server
	// default.
	transaction, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer transaction.Rollback()

	// set_config with is_local true rather than SET LOCAL, because SET takes no
	// query parameters. It takes the value as text, hence the formatting.
	threshold := strconv.FormatFloat(narrowingThreshold, 'f', -1, 64)
	if _, err := transaction.ExecContext(ctx, "SELECT set_config('pg_trgm.word_similarity_threshold', $1, true)", threshold); err != nil {
		return nil, err
	}

	rows, err := transaction.QueryContext(ctx, `
		WITH candidates AS (
			SELECT DISTINCT i.recipe_id
			FROM ingredients i
			JOIN unnest($1::text[]) AS token ON token <% i.name_norm
			ORDER BY i.recipe_id
			LIMIT $2
		)
		SELECT r.id, r.name, r.category, r.image_key, r.calories, r.total_cook_time, i.name
		FROM candidates c
		JOIN recipes r ON r.id = c.recipe_id
		JOIN ingredients i ON i.recipe_id = r.id
		ORDER BY r.id, i.created_at, i.id
	`, tokens, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candidates := make([]models.MatchCandidate, 0)
	for rows.Next() {
		var recipe models.RecipeSummary
		var ingredientName string
		if err := rows.Scan(
			&recipe.ID,
			&recipe.Name,
			&recipe.Category,
			&recipe.ImageKey,
			&recipe.Calories,
			&recipe.TotalCookTime,
			&ingredientName,
		); err != nil {
			return nil, err
		}

		// The join returns one row per Ingredient, ordered by Recipe, so a new
		// id starts a new candidate.
		if len(candidates) == 0 || candidates[len(candidates)-1].Recipe.ID != recipe.ID {
			recipe.ResolveImage(r.images)
			candidates = append(candidates, models.MatchCandidate{Recipe: recipe})
		}
		last := &candidates[len(candidates)-1]
		last.Ingredients = append(last.Ingredients, ingredientName)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := transaction.Commit(); err != nil {
		return nil, err
	}

	return candidates, nil
}
