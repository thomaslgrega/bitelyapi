package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/thomaslgrega/bitelyapi/internal/models"
)

type RecipeRepository struct {
	db *sql.DB
}

func NewRecipeRepository(db *sql.DB) *RecipeRepository {
	return &RecipeRepository{db: db}
}

func (r *RecipeRepository) GetRecipeById(ctx context.Context, id string) (models.Recipe, error) {
	row := r.db.QueryRowContext(
		ctx,
		"SELECT id, name, category, instructions, thumbnail_url, calories, total_cook_time FROM recipes WHERE id = $1",
		id,
	)

	var recipe models.Recipe
	if err := row.Scan(
		&recipe.ID,
		&recipe.Name,
		&recipe.Category,
		&recipe.Instructions,
		&recipe.ThumbnailUrl,
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

	return recipe, nil
}

func (r *RecipeRepository) GetRecipesByCategory(ctx context.Context, category string) ([]models.RecipeSummary, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, name, calories, thumbnail_url FROM recipes WHERE category = $1",
		category,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipes []models.RecipeSummary
	for rows.Next() {
		var recipe models.RecipeSummary
		if err := rows.Scan(&recipe.ID, &recipe.Name, &recipe.Calories, &recipe.ThumbnailUrl); err != nil {
			return nil, err
		}
		recipes = append(recipes, recipe)
	}

	return recipes, nil
}

func (r *RecipeRepository) CreateRecipe(ctx context.Context, input models.CreateRecipeInput) (*models.Recipe, error) {
	transaction, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer transaction.Rollback()

	var recipeID string
	err = transaction.QueryRowContext(ctx, `
		INSERT INTO recipes (name, category, instructions, thumbnail_url, calories, total_cook_time)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, input.Name, input.Category, input.Instructions, input.ThumbnailUrl, input.Calories, input.TotalCookTime).Scan(&recipeID)

	if err != nil {
		return nil, err
	}

	for _, ingredient := range input.Ingredients {
		_, err := transaction.ExecContext(ctx, `
			INSERT INTO ingredients (recipe_id, name, measurement)
			VALUES ($1, $2, $3)
		`, recipeID, ingredient.Name, ingredient.Measurement)

		if err != nil {
			return nil, err
		}
	}

	if err := transaction.Commit(); err != nil {
		return nil, err
	}

	return &models.Recipe{
		ID:            recipeID,
		Name:          input.Name,
		Category:      input.Category,
		Instructions:  input.Instructions,
		Ingredients:   input.Ingredients,
		ThumbnailUrl:  input.ThumbnailUrl,
		Calories:      input.Calories,
		TotalCookTime: input.TotalCookTime,
	}, nil
}
