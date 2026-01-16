package models

type RecipeSummary struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ThumbnailUrl string `json:"thumbnail_url,omitempty"`
	Calories     int    `json:"calories,omitempty"`
	// ImageData - might have to use S3? if so, have to think about when users are offline. Save image as data when bookmarked?
}
