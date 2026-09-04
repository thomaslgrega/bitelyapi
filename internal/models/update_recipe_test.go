package models

import (
	"encoding/json"
	"testing"
)

// The image left the recipe write, so an update decodes an image_key only to
// let the handler refuse it by name (ADR-0006).
func TestUpdateRecipeInputDecodesAnImageKeyOnlyToRefuseIt(t *testing.T) {
	var input UpdateRecipeInput
	if err := json.Unmarshal([]byte(`{"name":"Shakshuka","image_key":"incoming/abc"}`), &input); err != nil {
		t.Fatalf("failed to unmarshal input: %v", err)
	}

	if input.ImageKey != "incoming/abc" {
		t.Fatalf("expected the image key to decode, got %q", input.ImageKey)
	}
}

// The id lives in the path, so a body carrying one names nothing the update reads.
func TestUpdateRecipeInputDecodesNoID(t *testing.T) {
	var input UpdateRecipeInput
	if err := json.Unmarshal([]byte(`{"id":"recipe-from-body","name":"Shakshuka"}`), &input); err != nil {
		t.Fatalf("failed to unmarshal input: %v", err)
	}

	if input.ID != "" {
		t.Fatalf("expected the body id to be ignored, got %q", input.ID)
	}
}

func TestUpdateRecipeInputTrimNames(t *testing.T) {
	input := UpdateRecipeInput{
		Name:        "  Chicken Dinner ",
		Ingredients: []Ingredient{{ID: "ingredient-1", Name: "  Chicken Breast ", Measurement: "2 lbs"}},
	}

	input.TrimNames()

	if input.Name != "Chicken Dinner" {
		t.Fatalf("expected trimmed recipe name, got %q", input.Name)
	}
	if input.Ingredients[0].Name != "Chicken Breast" {
		t.Fatalf("expected trimmed ingredient name, got %q", input.Ingredients[0].Name)
	}
	if input.Ingredients[0].ID != "ingredient-1" {
		t.Fatalf("expected ingredient id to be left alone, got %q", input.Ingredients[0].ID)
	}
}
