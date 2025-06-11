package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func NewRepository() (*pgx.Conn, error) {
	log.Print("Connecting to database...")

	conn, err := pgx.Connect(context.Background(), "postgres://:@localhost:5432/elindor")
	if err != nil {
		return nil, fmt.Errorf("connecting to database failed: %w", err)
	}

	log.Print("Connected to database")

	return conn, nil
}
