package service

import (
	"context"
	"github.com/google/uuid"

	"elindor/repository"
)

type OrderService interface {
	CreateOrder(ctx context.Context, email string) (uuid.UUID, error)
	AddCandlesToOrder(ctx context.Context, orderID uuid.UUID, candleID uuid.UUID, quantity int) error
	UpdatePayedOrder(ctx context.Context, orderID uuid.UUID, sessionID string) error
}

type orderService struct {
	repo *repository.Repository
}

func NewOrderService(repo *repository.Repository) OrderService {
	return &orderService{
		repo: repo,
	}
}

func (os *orderService) CreateOrder(ctx context.Context, email string) (uuid.UUID, error) {
	conn, err := os.repo.GetConnection()
	if err != nil {
		return uuid.Nil, err
	}

	orderID, err := repository.CreateOrder(conn, email)
	if err != nil {
		return uuid.Nil, err
	}

	return orderID, nil
}

func (os *orderService) AddCandlesToOrder(ctx context.Context, orderID uuid.UUID, candleID uuid.UUID, quantity int) error {
	conn, err := os.repo.GetConnection()
	if err != nil {
		return err
	}

	err = repository.AddCandlesToOrder(conn, orderID, candleID, quantity)
	if err != nil {
		return err
	}

	return nil
}

func (os *orderService) UpdatePayedOrder(ctx context.Context, orderID uuid.UUID, sessionID string) error {
	conn, err := os.repo.GetConnection()
	if err != nil {
		return err
	}

	err = repository.UpdatePayedOrder(conn, orderID, sessionID)
	if err != nil {
		return err
	}

	return nil
}
