package repository

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"

	"elindor/domain"
	"elindor/handler/response"
)

func GetCandles(conn *pgx.Conn) ([]domain.Candle, error) {
	result, err := conn.Query(context.Background(), "SELECT id, name_hu, name_en, price_huf, price_eur, price_czk, img_url FROM data.candles")

	if err != nil {
		log.Printf("query failed: %v", err)
		return []domain.Candle{}, response.InternalServerError{
			Message: "Internal server error, please contact support with request ID",
		}
	}

	var candles []domain.Candle
	for result.Next() {
		var c domain.Candle
		err := result.Scan(&c.ID, &c.NameHU, &c.NameEN, &c.PriceHUF, &c.PriceEUR, &c.PriceCZK, &c.ImageURL)
		if err != nil {
			log.Printf("scan failed: %v", err)
			return nil, response.InternalServerError{
				Message: "Internal server error, please contact support with request ID",
			}
		}
		candles = append(candles, c)
	}

	return candles, nil
}

func GetCandleByID(conn *pgx.Conn, id string) (*domain.Candle, error) {
	result := conn.QueryRow(context.Background(), "SELECT * FROM data.candles WHERE id = $1", id)

	var c domain.Candle
	err := result.Scan(&c.ID, &c.NameHU, &c.NameEN, &c.PriceHUF, &c.PriceEUR, &c.DescriptionHU, &c.ImageURL, &c.PriceCZK, &c.DescriptionEN, &c.DescriptionCZ)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Candle not found
		}
		log.Printf("scan failed: %v", err)
		return nil, response.InternalServerError{
			Message: "Internal server error, please contact support with request ID",
		}
	}

	return &c, nil
}
