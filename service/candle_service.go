package service

import (
	"context"

	"elindor/domain"
	"elindor/repository"
)

type CandleService interface {
	GetCandles(ctx context.Context) ([]domain.Candle, error)
}

type candleService struct {
	repo *repository.Repository
}

func NewCandleService(repo *repository.Repository) CandleService {
	return &candleService{
		repo: repo,
	}
}

func (cs *candleService) GetCandles(ctx context.Context) ([]domain.Candle, error) {
	conn, err := cs.repo.GetConnection()
	if err != nil {
		return nil, err
	}

	candles, err := repository.GetCandles(conn)
	if err != nil {
		return nil, err
	}

	return candles, nil
}
