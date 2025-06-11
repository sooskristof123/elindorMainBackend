package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"elindor/domain"
)

func GetCandles(conn *pgx.Conn) ([]domain.Candle, error) {
	result, err := conn.Query(context.Background(), "SELECT * FROM elindor.candles")

	if err != nil {
		return []domain.Candle{}, fmt.Errorf("could not get candles: %w", err)
	}

	var candles []domain.Candle
	for result.Next() {
		var c domain.Candle
		err := result.Scan(&c.ID, &c.Name, &c.Scent, &c.Price)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		candles = append(candles, c)
	}

	return candles, nil
}
