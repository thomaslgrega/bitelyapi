package models

type CreateIngredientInput struct {
	Name        string `json:"name"`
	Measurement string `json:"measurement"`
}

type CreateRecipeInput struct {
	Name          string                  `json:"name"`
	Category      string                  `json:"category"`
	Instructions  string                  `json:"instructions,omitempty"`
	ThumbnailUrl  string                  `json:"thumbnail_url,omitempty"`
	Ingredients   []CreateIngredientInput `json:"ingredients"`
	Calories      int                     `json:"calories,omitempty"`
	TotalCookTime int                     `json:"total_cook_time,omitempty"`
}
