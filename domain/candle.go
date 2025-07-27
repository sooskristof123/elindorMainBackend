package domain

type Candle struct {
	ID       string  `json:"id"`
	NameHU   string  `json:"name_hu"`
	NameEN   string  `json:"name_en"`
	PriceHUF float64 `json:"price_huf"`
	PriceEUR float64 `json:"price_eur"`
}
