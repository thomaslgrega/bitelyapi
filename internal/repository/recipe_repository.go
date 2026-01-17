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

	_, err = transaction.ExecContext(ctx, `
		INSERT INTO recipes (id, name, category, instructions, thumbnail_url, calories, total_cook_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, input.ID, input.Name, input.Category, input.Instructions, input.ThumbnailUrl, input.Calories, input.TotalCookTime)
	if err != nil {
		return nil, err
	}

	for _, ingredient := range input.Ingredients {
		_, err := transaction.ExecContext(ctx, `
			INSERT INTO ingredients (id, recipe_id, name, measurement)
			VALUES ($1, $2, $3, $4)
		`, ingredient.ID, input.ID, ingredient.Name, ingredient.Measurement)
		if err != nil {
			return nil, err
		}
	}

	if err := transaction.Commit(); err != nil {
		return nil, err
	}

	return &models.Recipe{
		ID:            input.ID,
		Name:          input.Name,
		Category:      input.Category,
		Instructions:  input.Instructions,
		Ingredients:   input.Ingredients,
		ThumbnailUrl:  input.ThumbnailUrl,
		Calories:      input.Calories,
		TotalCookTime: input.TotalCookTime,
	}, nil
}

func (r *RecipeRepository) DeleteRecipe(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM recipes WHERE id = $1", id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *RecipeRepository) UpdateRecipe(ctx context.Context, recipe models.Recipe) error {
	transaction, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer transaction.Rollback()

	result, err := transaction.ExecContext(ctx, `
		UPDATE recipes
		SET name = $1, category = $2, instructions = $3, thumbnail_url = $4, calories = $5, total_cook_time = $6
		WHERE id = $7
	`, recipe.Name, recipe.Category, recipe.Instructions, recipe.ThumbnailUrl, recipe.Calories, recipe.TotalCookTime, recipe.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	_, err = transaction.ExecContext(ctx, "DELETE FROM ingredients WHERE recipe_id = $1", recipe.ID)
	if err != nil {
		return err
	}

	for _, ingredient := range recipe.Ingredients {
		_, err := transaction.ExecContext(ctx, `
			INSERT INTO ingredients (id, recipe_id, name, measurement)
			VALUES ($1, $2, $3, $4)
		`, ingredient.ID, recipe.ID, ingredient.Name, ingredient.Measurement)

		if err != nil {
			return err
		}
	}

	if err := transaction.Commit(); err != nil {
		return err
	}

	return nil
}