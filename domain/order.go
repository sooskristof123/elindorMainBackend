package domain

import "github.com/google/uuid"

type Address struct {
	Country string `json:"country"`
	City    string `json:"city"`
	Zip     string `json:"zip"`
	Street  string `json:"street"`
	Line1   string `json:"line1"`
}

type Order struct {
	ID             uuid.UUID `json:"id"`
	Email          string    `json:"email"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	Status         string    `json:"status"`
	PickupPoint    *string   `json:"pickup_point,omitempty"`
	IsHomeDelivery bool      `json:"is_home_delivery"`
	Country        *string   `json:"country,omitempty"`
	City           *string   `json:"city,omitempty"`
	Zipcode        *string   `json:"zipcode,omitempty"`
	Street         *string   `json:"street,omitempty"`
	Line1          *string   `json:"line1,omitempty"`
	SessionID      *string   `json:"session_id,omitempty"`
	PromotionID    *string   `json:"promotion_id,omitempty"`
}

type OrderCandle struct {
	OrderID  uuid.UUID `json:"order_id"`
	CandleID uuid.UUID `json:"candle_id"`
	Quantity int       `json:"quantity"`
	Candle   Candle    `json:"candle"`
}

type OrderWithCandles struct {
	Order   Order         `json:"order"`
	Candles []OrderCandle `json:"candles"`
}
