package models

import (
	"encoding/json"
	"testing"
)

func TestImageLocatorURLFor(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		key     string
		want    string
	}{
		{
			name:    "composes the public URL from the key",
			baseURL: "https://img.bitely.app",
			key:     "recipes/abc/image.jpg",
			want:    "https://img.bitely.app/recipes/abc/image.jpg",
		},
		{
			name:    "a trailing slash on the base does not double",
			baseURL: "https://img.bitely.app/",
			key:     "recipes/abc/image.jpg",
			want:    "https://img.bitely.app/recipes/abc/image.jpg",
		},
		{
			name:    "no key means no URL, not a bare hostname",
			baseURL: "https://img.bitely.app",
			key:     "",
			want:    "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := NewImageLocator(test.baseURL).URLFor(test.key)
			if got != test.want {
				t.Fatalf("expected %q, got %q", test.want, got)
			}
		})
	}
}

// A Recipe scans a stored key and answers a fetchable URL. The key stays on
// the struct and off the wire: nothing decodes a Recipe, so it is tagged out
// of JSON rather than blanked on the way past (ADR-0006).
func TestRecipeResolveImage(t *testing.T) {
	recipe := Recipe{ID: "recipe-1", Name: "Shakshuka", ImageKey: "recipes/recipe-1/image.jpg"}

	recipe.ResolveImage(NewImageLocator("https://img.bitely.app"))

	if recipe.ImageUrl != "https://img.bitely.app/recipes/recipe-1/image.jpg" {
		t.Fatalf("expected composed image url, got %q", recipe.ImageUrl)
	}
	if recipe.ImageKey != "recipes/recipe-1/image.jpg" {
		t.Fatalf("expected the key to survive resolving, got %q", recipe.ImageKey)
	}

	body, err := json.Marshal(recipe)
	if err != nil {
		t.Fatalf("failed to marshal recipe: %v", err)
	}

	var encoded map[string]any
	if err := json.Unmarshal(body, &encoded); err != nil {
		t.Fatalf("failed to unmarshal recipe: %v", err)
	}
	if _, present := encoded["image_key"]; present {
		t.Fatalf("expected no image_key in the response, got %s", body)
	}
	if encoded["image_url"] != "https://img.bitely.app/recipes/recipe-1/image.jpg" {
		t.Fatalf("expected image_url in the response, got %s", body)
	}
}

func TestRecipeResolveImageWithoutAnImage(t *testing.T) {
	recipe := Recipe{ID: "recipe-1", Name: "Shakshuka"}

	recipe.ResolveImage(NewImageLocator("https://img.bitely.app"))

	body, err := json.Marshal(recipe)
	if err != nil {
		t.Fatalf("failed to marshal recipe: %v", err)
	}

	var encoded map[string]any
	if err := json.Unmarshal(body, &encoded); err != nil {
		t.Fatalf("failed to unmarshal recipe: %v", err)
	}
	if _, present := encoded["image_url"]; present {
		t.Fatalf("expected an imageless recipe to carry no image_url, got %s", body)
	}
}

func TestRecipeSummaryResolveImage(t *testing.T) {
	summary := RecipeSummary{ID: "recipe-1", Name: "Shakshuka", ImageKey: "recipes/recipe-1/image.jpg"}

	summary.ResolveImage(NewImageLocator("https://img.bitely.app"))

	if summary.ImageUrl != "https://img.bitely.app/recipes/recipe-1/image.jpg" {
		t.Fatalf("expected composed image url, got %q", summary.ImageUrl)
	}
	if summary.ImageKey != "recipes/recipe-1/image.jpg" {
		t.Fatalf("expected the key to survive resolving, got %q", summary.ImageKey)
	}

	body, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("failed to marshal summary: %v", err)
	}

	var encoded map[string]any
	if err := json.Unmarshal(body, &encoded); err != nil {
		t.Fatalf("failed to unmarshal summary: %v", err)
	}
	if _, present := encoded["image_key"]; present {
		t.Fatalf("expected no image_key in the response, got %s", body)
	}
}

// The image left the recipe write, so a key sent to a Recipe reaches no field
// at all (ADR-0006).
func TestRecipeDecodesNoImageKey(t *testing.T) {
	var recipe Recipe
	if err := json.Unmarshal([]byte(`{"name":"Shakshuka","image_key":"incoming/abc"}`), &recipe); err != nil {
		t.Fatalf("failed to unmarshal recipe: %v", err)
	}

	if recipe.ImageKey != "" {
		t.Fatalf("expected the image key to be ignored, got %q", recipe.ImageKey)
	}
}

func TestCreateRecipeInputDecodesAnImageKey(t *testing.T) {
	var input CreateRecipeInput
	if err := json.Unmarshal([]byte(`{"name":"Shakshuka","image_key":"incoming/abc"}`), &input); err != nil {
		t.Fatalf("failed to unmarshal input: %v", err)
	}

	if input.ImageKey != "incoming/abc" {
		t.Fatalf("expected the image key to decode, got %q", input.ImageKey)
	}
}
