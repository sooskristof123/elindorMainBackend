package domain

type Candle struct {
	ID    string  `json:"-"`
	Name  string  `json:"name"`
	Scent string  `json:"scent"`
	Price float64 `json:"price"`
}
