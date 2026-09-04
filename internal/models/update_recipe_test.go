package models

import (
	"encoding/json"
	"testing"
)

// The image left the recipe write, so an update decodes an image_key only to
// let the handler refuse it by name. Presence is what it records: "" and null
// are what a client asking for the old absent-means-delete sends (ADR-0006).
func TestUpdateRecipeInputRecordsThatAnImageKeyWasSent(t *testing.T) {
	tests := []struct {
		name string
		body string
		sent bool
	}{
		{name: "a staged key", body: `{"name":"Shakshuka","image_key":"incoming/abc"}`, sent: true},
		{name: "an empty key", body: `{"name":"Shakshuka","image_key":""}`, sent: true},
		{name: "an explicit null", body: `{"name":"Shakshuka","image_key":null}`, sent: true},
		{name: "no image field at all", body: `{"name":"Shakshuka"}`, sent: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var input UpdateRecipeInput
			if err := json.Unmarshal([]byte(test.body), &input); err != nil {
				t.Fatalf("failed to unmarshal input: %v", err)
			}

			if sent := input.ImageKey != nil; sent != test.sent {
				t.Fatalf("expected sent to be %v, got %v", test.sent, sent)
			}
		})
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
