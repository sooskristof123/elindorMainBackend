package repository

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"

	"elindor/domain"
	"elindor/handler/response"
)

func GetCollections(conn *pgx.Conn) ([]domain.Collection, error) {
	result, err := conn.Query(context.Background(), "SELECT * FROM data.collections")

	if err != nil {
		log.Printf("query failed: %v", err)
		return nil, response.InternalServerError{
			Message: "Internal server error, please contact support with request ID",
		}
	}

	var collections []domain.Collection
	for result.Next() {
		var c domain.Collection
		err := result.Scan(&c.ID, &c.Name, &c.Description)
		if err != nil {
			log.Printf("scan failed: %v", err)
			return nil, response.InternalServerError{
				Message: "Internal server error, please contact support with request ID",
			}
		}
		collections = append(collections, c)
	}

	return collections, nil
}

func GetCollection(conn *pgx.Conn, name string) ([]domain.Candle, error) {
	result, err := conn.Query(context.Background(),
		`SELECT *
         FROM data.candles
         WHERE id IN (
             SELECT data.collections_candles.candles_id
             FROM data.collections_candles
             WHERE data.collections_candles.collections_name = $1
        )`,
		name)

	if err != nil {
		log.Printf("query failed: %v", err)
		return nil, response.InternalServerError{
			Message: "Internal server error, please contact support with request ID",
		}
	}

	var candles []domain.Candle
	for result.Next() {
		var c domain.Candle
		err := result.Scan(&c.ID, &c.NameHU, &c.NameEN, &c.PriceHUF, &c.PriceEUR)

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
