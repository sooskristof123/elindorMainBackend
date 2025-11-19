package service

import (
	"context"
	"errors"

	"elindor/domain"
	"elindor/repository"
)

type PromotionService interface {
	GetPromotionByName(ctx context.Context, promotionName string, email string) (*domain.PromotionResponse, error)
	SavePromotionUsage(ctx context.Context, promotionID string, email string) error
}

type promotionService struct {
	repo *repository.Repository
}

func NewPromotionService(repo *repository.Repository) PromotionService {
	return &promotionService{
		repo: repo,
	}
}

func (ps *promotionService) GetPromotionByName(ctx context.Context, promotionName string, email string) (*domain.PromotionResponse, error) {
	conn, err := ps.repo.GetConnection()
	if err != nil {
		return nil, err
	}

	promotion, err := repository.GetPromotionByName(conn, promotionName)
	if err != nil {
		return nil, err
	}

	if promotion == nil {
		return nil, errors.New("promotion not found")
	}

	available, err := repository.GetAvailablePromotionByNameAndEmail(conn, email, promotion.ID)
	if err != nil {
		return nil, err
	}

	return &domain.PromotionResponse{Promotion: *promotion, Available: available}, nil
}

func (ps *promotionService) SavePromotionUsage(ctx context.Context, promotionID string, email string) error {
	conn, err := ps.repo.GetConnection()
	if err != nil {
		return err
	}

	return repository.SavePromotionUsage(conn, promotionID, email)
}
