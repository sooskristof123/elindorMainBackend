package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"

	"elindor/handler/response"
)

type Repository struct {
	Conn *pgx.Conn
}

func NewRepository() (*Repository, error) {
	log.Print("Connecting to database...")

	conn, err := pgx.Connect(context.Background(), "postgres://adagadt:60GGq727@79.139.57.135:5432/elindor")
	if err != nil {
		return nil, fmt.Errorf("connecting to database failed: %w", err)
	}

	log.Print("Connected to database")

	return &Repository{
		Conn: conn,
	}, nil
}

func (r *Repository) GetConnection() (*pgx.Conn, error) {
	if err := r.Conn.Ping(context.Background()); err != nil {
		log.Print("Reconnecting to database...")
		conn, err := pgx.Connect(context.Background(), "postgres://dagadt:60GGq727@localhost:5432/elindor")

		if err != nil {
			log.Printf("reconnecting to database failed: %v", err)
			return nil, response.InternalServerError{
				Message: "Internal server error, please contact support with request ID",
			}
		}

		r.Conn = conn
	}

	return r.Conn, nil
}
