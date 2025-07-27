package service

import (
	"context"

	"elindor/domain"
	"elindor/repository"
)

type CollectionService interface {
	GetCollections(ctx context.Context) ([]domain.Collection, error)
	GetCollection(ctx context.Context, name string) ([]domain.Candle, error)
}

type collectionService struct {
	repo *repository.Repository
}

func NewCollectionService(repo *repository.Repository) CollectionService {
	return &collectionService{
		repo: repo,
	}
}

func (cs *collectionService) GetCollections(ctx context.Context) ([]domain.Collection, error) {
	conn, err := cs.repo.GetConnection()
	if err != nil {
		return nil, err
	}

	collections, err := repository.GetCollections(conn)
	if err != nil {
		return nil, err
	}

	return collections, nil
}

func (cs *collectionService) GetCollection(ctx context.Context, name string) ([]domain.Candle, error) {
	conn, err := cs.repo.GetConnection()
	if err != nil {
		return nil, err
	}

	candles, err := repository.GetCollection(conn, name)
	if err != nil {
		return nil, err
	}

	return candles, nil
}
