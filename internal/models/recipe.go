package models

import "strings"

type Recipe struct {
	ID            string       `json:"id"`
	UserID        string       `json:"user_id"`
	Name          string       `json:"name"`
	Category      string       `json:"category"`
	Instructions  string       `json:"instructions,omitempty"`
	ImageKey      string       `json:"image_key,omitempty"`
	ImageUrl      string       `json:"image_url,omitempty"`
	Ingredients   []Ingredient `json:"ingredients"`
	Calories      int          `json:"calories,omitempty"`
	TotalCookTime int          `json:"total_cook_time,omitempty"`
}

// TrimNames trims the names a client sent so the value validated, the value
// stored, and the value returned are all the same string.
func (r *Recipe) TrimNames() {
	r.Name = strings.TrimSpace(r.Name)
	for i := range r.Ingredients {
		r.Ingredients[i].Name = strings.TrimSpace(r.Ingredients[i].Name)
	}
}
