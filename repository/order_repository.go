package repository

import (
	"context"
	"elindor/domain"
	"github.com/google/uuid"
	"log"

	"github.com/jackc/pgx/v5"

	"elindor/handler/response"
)

func CreateOrder(conn *pgx.Conn, email string) (uuid.UUID, error) {
	orderID := uuid.New()
	_, err := conn.Exec(context.Background(),
		`INSERT INTO data.orders (id, email, status, created_at) VALUES ($1, $2, $3, NOW())`, orderID, email, "pending_payment")

	if err != nil {
		log.Printf("query failed: %v", err)
		return uuid.Nil, response.InternalServerError{
			Message: "Internal server error, please contact support with request ID",
		}
	}

	return orderID, nil
}

func AddCandlesToOrder(conn *pgx.Conn, orderID uuid.UUID, candleID uuid.UUID, quantity int) error {
	_, err := conn.Exec(context.Background(), "INSERT INTO data.order_candles (order_id, candle_id, quantity) VALUES ($1, $2, $3)", orderID, candleID, quantity)
	if err != nil {
		log.Printf("query failed: %v", err)
		return response.InternalServerError{
			Message: "Internal server error, please contact support with request ID",
		}
	}

	return nil
}

func AddPickUpPointToOrder(conn *pgx.Conn, orderID uuid.UUID, pickUpPoint string) error {
	_, err := conn.Exec(context.Background(), "UPDATE data.orders SET pickup_point = $1, is_homedelivery = false WHERE id = $2", pickUpPoint, orderID)
	if err != nil {
		log.Printf("query failed: %v", err)
		return response.InternalServerError{
			Message: "Internal server error, please contact support with request ID",
		}
	}

	return nil
}

func AddAddressToOrder(conn *pgx.Conn, orderID uuid.UUID, address domain.Address) error {
	_, err := conn.Exec(
		context.Background(),
		"UPDATE data.orders SET is_homedelivery = true, country = $1, city = $2, zipcode = $3, street = $4, line1 = $5 WHERE id = $6",
		address.Country, address.City, address.Zip, address.Street, address.Line1, orderID,
	)
	if err != nil {
		log.Printf("query failed: %v", err)
		return response.InternalServerError{
			Message: "Internal server error, please contact support with request ID",
		}
	}

	return nil
}

func UpdatePayedOrder(conn *pgx.Conn, orderID uuid.UUID, sessionID string) error {
	_, err := conn.Exec(context.Background(), "UPDATE data.orders SET status = $1, session_id = $3, paid_at = NOW() WHERE id = $2", "paid", orderID, sessionID)
	if err != nil {
		log.Printf("query failed: %v", err)
		return response.InternalServerError{
			Message: "Internal server error, please contact support with request ID",
		}
	}

	return nil
}
