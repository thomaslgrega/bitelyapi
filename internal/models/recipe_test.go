package models

import "testing"

func TestRecipeTrimNames(t *testing.T) {
	recipe := Recipe{
		Name:        "  Chicken Dinner ",
		Ingredients: []Ingredient{{ID: "ingredient-1", Name: "  Chicken Breast ", Measurement: "2 lbs"}},
	}

	recipe.TrimNames()

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
