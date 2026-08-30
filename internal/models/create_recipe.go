package models

import "strings"

type CreateIngredientInput struct {
	Name        string `json:"name"`
	Measurement string `json:"measurement"`
}

type CreateRecipeInput struct {
	Name          string                  `json:"name"`
	Category      string                  `json:"category"`
	Instructions  string                  `json:"instructions,omitempty"`
	ImageKey      string                  `json:"image_key,omitempty"`
	Ingredients   []CreateIngredientInput `json:"ingredients"`
	Calories      int                     `json:"calories,omitempty"`
	TotalCookTime int                     `json:"total_cook_time,omitempty"`
}

// TrimNames trims the names a client sent so the value validated, the value
// stored, and the value returned are all the same string.
func (input *CreateRecipeInput) TrimNames() {
	input.Name = strings.TrimSpace(input.Name)
	for i := range input.Ingredients {
		input.Ingredients[i].Name = strings.TrimSpace(input.Ingredients[i].Name)
	}
}
