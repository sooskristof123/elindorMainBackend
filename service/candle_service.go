package service

import (
	"context"

	"elindor/domain"
)

type CandleService interface {
	GetCandles(ctx context.Context) []domain.Candle
}

type candleService struct {
}

func NewCandleService() CandleService {
	return &candleService{}
}

func (cs *candleService) GetCandles(ctx context.Context) []domain.Candle {
	var candles = []domain.Candle{
		{
			ID:   "1",
			Name: "<UNK>",
		},
		{
			ID:   "2",
			Name: "<UNK>",
		},
	}

	return candles
}
