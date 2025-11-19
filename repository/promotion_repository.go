package repository

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"

	"elindor/domain"
	"elindor/handler/response"
)

func GetPromotionByName(conn *pgx.Conn, name string) (*domain.Promotion, error) {
	result := conn.QueryRow(context.Background(), "SELECT * FROM data.promotions WHERE name = $1", name)

	var p domain.Promotion
	err := result.Scan(&p.ID, &p.Name, &p.Percentage)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Promotion not found
		}
		log.Printf("scan failed: %v", err)
		return nil, response.InternalServerError{
			Message: "Internal server error, please contact support with request ID",
		}
	}

	return &p, nil
}

func GetAvailablePromotionByNameAndEmail(conn *pgx.Conn, email string, promotionID string) (bool, error) {
	result := conn.QueryRow(context.Background(), "SELECT 1 FROM data.promotions_users WHERE promotion_id = $1 and email = $2", promotionID, email)

	var exists int
	err := result.Scan(&exists)
	if err != nil {
		if err == pgx.ErrNoRows {
			return true, nil // Promotion not found in promotions_users = available for this user
		}
		log.Printf("scan failed: %v", err)
		return false, response.InternalServerError{
			Message: "Internal server error, please contact support with request ID",
		}
	}

	// If we found a row, the promotion has already been used by this email
	return false, nil
}

func SavePromotionUsage(conn *pgx.Conn, promotionID string, email string) error {
	_, err := conn.Exec(context.Background(),
		"INSERT INTO data.promotions_users (promotion_id, email) VALUES ($1, $2)",
		promotionID, email)

	if err != nil {
		log.Printf("query failed: %v", err)
		return response.InternalServerError{
			Message: "Internal server error, please contact support with request ID",
		}
	}

	return nil
}
