package domain

type Collection struct {
	ID          string `json:"-"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
