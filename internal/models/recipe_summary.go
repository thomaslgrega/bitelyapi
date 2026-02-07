package models

type RecipeSummary struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Category string `json:"category"`
	ThumbnailUrl string `json:"thumbnail_url,omitempty"`
	Calories     int    `json:"calories,omitempty"`
	TotalCookTime int `json:"total_cook_time,omitempty"`
}
