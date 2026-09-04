package models

import (
	"encoding/json"
	"strings"
)

// UpdateRecipeInput is what PUT /recipes/{id} writes. It carries no image: a
// Recipe Image is written through its own sub-resource (ADR-0006).
type UpdateRecipeInput struct {
	// ID comes from the path, so a body naming another Recipe changes nothing.
	ID            string       `json:"-"`
	Name          string       `json:"name"`
	Category      string       `json:"category"`
	Instructions  string       `json:"instructions,omitempty"`
	Ingredients   []Ingredient `json:"ingredients"`
	Calories      int          `json:"calories,omitempty"`
	TotalCookTime int          `json:"total_cook_time,omitempty"`

	// ImageKey is decoded only so the write can refuse it by name, and raw so
	// that "" and null are refused too: a stale client asks for the deletion
	// this write no longer performs by sending exactly those.
	ImageKey json.RawMessage `json:"image_key,omitempty"`
}

// TrimNames trims the names a client sent so the value validated, the value
// stored, and the value returned are all the same string.
func (input *UpdateRecipeInput) TrimNames() {
	input.Name = strings.TrimSpace(input.Name)
	for i := range input.Ingredients {
		input.Ingredients[i].Name = strings.TrimSpace(input.Ingredients[i].Name)
	}
}
