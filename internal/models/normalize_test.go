package models

import "testing"

func TestCreateRecipeInputNormalize(t *testing.T) {
	input := CreateRecipeInput{
		Name: "  Chicken Dinner ",
		Ingredients: []CreateIngredientInput{
			{Name: "\tChicken Breast\n", Measurement: "2 lbs"},
			{Name: "Salt", Measurement: "1 tsp"},
		},
	}

	input.Normalize()

	if input.Name != "Chicken Dinner" {
		t.Fatalf("expected trimmed recipe name, got %q", input.Name)
	}
	if input.Ingredients[0].Name != "Chicken Breast" {
		t.Fatalf("expected trimmed ingredient name, got %q", input.Ingredients[0].Name)
	}
	if input.Ingredients[0].Measurement != "2 lbs" {
		t.Fatalf("expected measurement to be left alone, got %q", input.Ingredients[0].Measurement)
	}
	if input.Ingredients[1].Name != "Salt" {
		t.Fatalf("expected an already-trimmed name to be unchanged, got %q", input.Ingredients[1].Name)
	}
}

func TestRecipeNormalize(t *testing.T) {
	recipe := Recipe{
		Name:        "  Chicken Dinner ",
		Ingredients: []Ingredient{{ID: "ingredient-1", Name: "  Chicken Breast ", Measurement: "2 lbs"}},
	}

	recipe.Normalize()

	if recipe.Name != "Chicken Dinner" {
		t.Fatalf("expected trimmed recipe name, got %q", recipe.Name)
	}
	if recipe.Ingredients[0].Name != "Chicken Breast" {
		t.Fatalf("expected trimmed ingredient name, got %q", recipe.Ingredients[0].Name)
	}
	if recipe.Ingredients[0].ID != "ingredient-1" {
		t.Fatalf("expected ingredient id to be left alone, got %q", recipe.Ingredients[0].ID)
	}
}
