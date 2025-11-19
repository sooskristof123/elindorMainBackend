package repository

import (
	"context"
	"elindor/domain"
	"log"
	"strconv"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"

	"elindor/handler/response"
)

func CreateOrder(conn *pgx.Conn, email, firstName, lastName string, phone *string, promotionID *string, totalPrice int64, discountedPrice *int64, shippingPrice int64, billingAddressMatch bool, billingCountry, billingCity, billingZip, billingStreet, billingLine1 *string) (uuid.UUID, error) {
	orderID := uuid.New()

	// Convert prices to strings
	totalPriceStr := strconv.FormatInt(totalPrice, 10)
	var discountedPriceStr *string
	if discountedPrice != nil {
		str := strconv.FormatInt(*discountedPrice, 10)
		discountedPriceStr = &str
	}
	shippingPriceStr := strconv.FormatInt(shippingPrice, 10)

	_, err := conn.Exec(context.Background(),
		`INSERT INTO data.orders (id, email, first_name, last_name, phone_number, promotion_id, total_price, discounted_price, shipping_price, billing_address_match, billing_country, billing_city, billing_zipcode, billing_street, billing_line1, status, created_at) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW())`,
		orderID, email, firstName, lastName, phone, promotionID, totalPriceStr, discountedPriceStr, shippingPriceStr, billingAddressMatch, billingCountry, billingCity, billingZip, billingStreet, billingLine1, "pending_payment")

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

func GetOrderWithCandles(conn *pgx.Conn, orderID uuid.UUID) (*domain.OrderWithCandles, error) {
	// First get the order details
	orderQuery := `
		SELECT id, email, first_name, last_name, status, pickup_point, is_homedelivery, country, city, zipcode, street, line1, session_id, promotion_id
		FROM data.orders 
		WHERE id = $1
	`

	var order domain.Order
	var pickupPoint, country, city, zipcode, street, line1, sessionID, promotionID *string

	err := conn.QueryRow(context.Background(), orderQuery, orderID).Scan(
		&order.ID, &order.Email, &order.FirstName, &order.LastName, &order.Status, &pickupPoint, &order.IsHomeDelivery,
		&country, &city, &zipcode, &street, &line1, &sessionID, &promotionID,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Order not found
		}
		log.Printf("query failed: %v", err)
		return nil, response.InternalServerError{
			Message: "Internal server error, please contact support with request ID",
		}
	}

	// Assign nullable fields
	order.PickupPoint = pickupPoint
	order.Country = country
	order.City = city
	order.Zipcode = zipcode
	order.Street = street
	order.Line1 = line1
	order.SessionID = sessionID
	order.PromotionID = promotionID

	// Get order candles with candle details
	candlesQuery := `
		SELECT oc.order_id, oc.candle_id, oc.quantity,
			   c.id, c.name_hu, c.name_en, c.price_huf, c.price_eur, c.price_czk, c.description_hu, c.img_url, c.description_en, c.description_cs
		FROM data.order_candles oc
		JOIN data.candles c ON oc.candle_id = c.id
		WHERE oc.order_id = $1
	`

	rows, err := conn.Query(context.Background(), candlesQuery, orderID)
	if err != nil {
		log.Printf("query failed: %v", err)
		return nil, response.InternalServerError{
			Message: "Internal server error, please contact support with request ID",
		}
	}
	defer rows.Close()

	var orderCandles []domain.OrderCandle
	for rows.Next() {
		var oc domain.OrderCandle
		var candle domain.Candle

		err := rows.Scan(
			&oc.OrderID, &oc.CandleID, &oc.Quantity,
			&candle.ID, &candle.NameHU, &candle.NameEN, &candle.PriceHUF, &candle.PriceEUR,
			&candle.PriceCZK, &candle.DescriptionHU, &candle.ImageURL, &candle.DescriptionEN, &candle.DescriptionCZ,
		)
		if err != nil {
			log.Printf("scan failed: %v", err)
			return nil, response.InternalServerError{
				Message: "Internal server error, please contact support with request ID",
			}
		}

		oc.Candle = candle
		orderCandles = append(orderCandles, oc)
	}

	return &domain.OrderWithCandles{
		Order:   order,
		Candles: orderCandles,
	}, nil
}
