package models

type Recipe struct {
	ID            string       `json:"id"`
	UserID        string       `json:"user_id"`
	Name          string       `json:"name"`
	Category      string       `json:"category"`
	Instructions  string       `json:"instructions,omitempty"`
	ThumbnailUrl  string       `json:"thumbnail_url,omitempty"`
	Ingredients   []Ingredient `json:"ingredients"`
	Calories      int          `json:"calories,omitempty"`
	TotalCookTime int          `json:"total_cook_time,omitempty"`
}
