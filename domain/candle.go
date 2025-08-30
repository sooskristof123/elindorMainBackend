package domain

import "github.com/google/uuid"

type Candle struct {
	ID            string  `json:"id"`
	NameHU        string  `json:"name_hu"`
	NameEN        string  `json:"name_en"`
	PriceHUF      float64 `json:"price_huf"`
	PriceEUR      float64 `json:"price_eur"`
	PriceCZK      float64 `json:"price_czk"`
	DescriptionHU string  `json:"description,omitempty"`
	DescriptionEN string  `json:"description_en,omitempty"`
	DescriptionCZ string  `json:"description_cz,omitempty"`
	ImageURL      string  `json:"image_url,omitempty"`
}

type CandleItem struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name" bson:"name"`
	Price    int64     `json:"price" bson:"price"`
	Lang     string    `json:"lang" bson:"lang"`
	Quantity int       `json:"quantity" bson:"quantity"`
}
