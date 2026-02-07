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

func (r *RecipeRepository) GetRecipesByUserID(ctx context.Context, userID string) ([]models.Recipe, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, category, thumbnail_url, calories, total_cook_time
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
			&recipe.ThumbnailUrl,
			&recipe.Calories,
			&recipe.TotalCookTime,
		); err != nil {
			return nil, err
		}
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
		"SELECT id, user_id, name, category, instructions, thumbnail_url, calories, total_cook_time FROM recipes WHERE id = $1",
		id,
	)

	var recipe models.Recipe
	if err := row.Scan(
		&recipe.ID,
		&recipe.UserID,
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
		"SELECT id, name, category, thumbnail_url, calories, total_cook_time FROM recipes WHERE category = $1",
		category,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recipes := make([]models.RecipeSummary, 0)
	for rows.Next() {
		var recipe models.RecipeSummary
		if err := rows.Scan(&recipe.ID, &recipe.Name, &recipe.Category, &recipe.ThumbnailUrl, &recipe.Calories, &recipe.TotalCookTime); err != nil {
			return nil, err
		}
		recipes = append(recipes, recipe)
	}

	return recipes, nil
}

func (r *RecipeRepository) CreateRecipe(ctx context.Context, userID string, input models.CreateRecipeInput) (*models.Recipe, error) {
	transaction, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer transaction.Rollback()

	var recipeID string
	err = transaction.QueryRowContext(ctx, `
		INSERT INTO recipes (user_id, name, category, instructions, thumbnail_url, calories, total_cook_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, userID, input.Name, input.Category, input.Instructions, input.ThumbnailUrl, input.Calories, input.TotalCookTime).Scan(&recipeID)
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
			ID: ingredientID,
			Name: ingredient.Name,
			Measurement: ingredient.Measurement,
		})
	}

	if err := transaction.Commit(); err != nil {
		return nil, err
	}

	return &models.Recipe{
		ID:            recipeID,
		UserID:        userID,
		Name:          input.Name,
		Category:      input.Category,
		Instructions:  input.Instructions,
		ThumbnailUrl:  input.ThumbnailUrl,
		Ingredients:   ingredients,
		Calories:      input.Calories,
		TotalCookTime: input.TotalCookTime,
	}, nil
}

func (r *RecipeRepository) DeleteRecipe(ctx context.Context, id string, userID string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM recipes WHERE id = $1 AND user_id = $2", id, userID)
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

	return nil
}

func (r *RecipeRepository) UpdateRecipe(ctx context.Context, recipe models.Recipe, userID string) error {
	transaction, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer transaction.Rollback()

	result, err := transaction.ExecContext(ctx, `
		UPDATE recipes
		SET name = $1, category = $2, instructions = $3, thumbnail_url = $4, calories = $5, total_cook_time = $6
		WHERE id = $7 AND user_id = $8
	`, recipe.Name, recipe.Category, recipe.Instructions, recipe.ThumbnailUrl, recipe.Calories, recipe.TotalCookTime, recipe.ID, userID)
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