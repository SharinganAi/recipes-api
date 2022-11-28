package models

type Recipe struct {
	Name         string   `json:"name"`
	Tags         []string `json:"tags"`
	Ingredients  []string `json:"ingredients"`
	Instructions []string `json:"instructions"`
	PublishDate  string   `json:"publishedAt"`
}
