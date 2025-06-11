package service

import (
	"context"

	"elindor/repository"
	"github.com/jackc/pgx/v5"

	"elindor/domain"
)

type CandleService interface {
	GetCandles(ctx context.Context) ([]domain.Candle, error)
}

type candleService struct {
	conn *pgx.Conn
}

func NewCandleService(conn *pgx.Conn) CandleService {
	return &candleService{
		conn: conn,
	}
}

func (cs *candleService) GetCandles(ctx context.Context) ([]domain.Candle, error) {
	candles, err := repository.GetCandles(cs.conn)

	if err != nil {
		return nil, err
	}

	return candles, nil
}
