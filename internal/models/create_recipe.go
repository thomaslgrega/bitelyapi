package models

type CreateRecipeInput struct {
	Name          string       `json:"name"`
	Category      string       `json:"category"`
	Instructions  string       `json:"instructions"`
	ThumbnailUrl  string       `json:"thumbnail_url"`
	Ingredients   []Ingredient `json:"ingredients"`
	Calories      int          `json:"calories"`
	TotalCookTime int          `json:"total_cook_time"`
}
