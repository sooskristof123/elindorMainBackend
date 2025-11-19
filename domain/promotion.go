package domain

type Promotion struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Percentage int    `json:"percentage"`
}

type PromotionResponse struct {
	Promotion Promotion `json:"promotion"`
	Available bool      `json:"available"`
}
