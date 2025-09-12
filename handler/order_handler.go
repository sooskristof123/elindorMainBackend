package handler

import (
	"context"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/webhook"

	"elindor/domain"
	"elindor/service"
)

type createOrderRequest struct {
	Currency    *string             `json:"currency"`
	Email       string              `json:"email" binding:"required,email"`
	Address     *domain.Address     `json:"address"`
	PickupPoint *string             `json:"pickup_point"`
	Candles     []domain.CandleItem `json:"candles" binding:"required"`
}

type OrderHandler struct {
	OrderService  service.OrderService
	CandleService service.CandleService
}

func (oh *OrderHandler) OrderEndpoints(router *gin.Engine) {
	router.POST("orders", oh.CreateOrder)
	router.PUT("orders/:order_id", oh.UpdatePayedOrder)
	router.POST("orders/webhook", oh.HandleStripeWebhook)
}

func NewOrderHandler(orderService service.OrderService, candleService service.CandleService) *OrderHandler {
	return &OrderHandler{
		orderService,
		candleService,
	}
}

func (oh *OrderHandler) CreateOrder(ctx *gin.Context) {
	var req createOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	// Validation
	if req.Address == nil && req.PickupPoint == nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Either address or pickup point must be provided",
		})
		return
	}
	if req.Address != nil {
		if req.Address.Country == "" || req.Address.City == "" || req.Address.Zip == "" || req.Address.Street == "" {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "Incomplete address information",
			})
			return
		}
	} else if req.PickupPoint != nil {
		if *req.PickupPoint == "" {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "Pickup point cannot be empty",
			})
			return
		}
	}

	// Candle price validation
	if req.Currency == nil || *req.Currency == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Currency is required",
		})
		return
	}
	currency := *req.Currency
	for _, candle := range req.Candles {
		candleData, err := oh.CandleService.GetCandleByID(ctx, candle.ID.String())
		if err != nil || candleData == nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "Invalid candle ID: " + candle.ID.String(),
			})
			return
		}
		var expectedPrice int64
		if currency == "huf" {
			expectedPrice = int64(math.Round(candleData.PriceHUF))
		} else if currency == "eur" {
			expectedPrice = int64(math.Round(candleData.PriceEUR))
		} else if currency == "czk" {
			expectedPrice = int64(math.Round(candleData.PriceCZK))
		} else {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "Unsupported currency: " + currency,
			})
			return
		}
		if candle.Price != expectedPrice {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "Candle price mismatch for candle ID: " + candle.ID.String(),
			})
			return
		}
	}

	// 1. Create order in DB
	orderID, err := oh.OrderService.CreateOrder(ctx, req.Email)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to create order: " + err.Error(),
		})
		return
	}

	// 2. Add items to order
	for _, candle := range req.Candles {
		err = oh.OrderService.AddCandlesToOrder(ctx, orderID, candle.ID, candle.Quantity)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to add candles to order: " + err.Error(),
			})
			return
		}
	}

	// 3. Add address to order
	if req.Address == nil {
		err = oh.OrderService.AddPickUpPointToOrder(ctx, orderID, *req.PickupPoint)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to add pickup point to order: " + err.Error(),
			})
			return
		}
	} else {
		err = oh.OrderService.AddAddressToOrder(ctx, orderID, *req.Address)
	}

	// 4. Convert candles into Stripe line items
	stripe.Key = os.Getenv("STRIPE_SECRET")
	var lineItems []*stripe.CheckoutSessionLineItemParams
	for _, candle := range req.Candles {
		lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String("huf"),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name: stripe.String(candle.Name), // candle struct should have Name
				},
				UnitAmount: stripe.Int64(candle.Price * 100),
			},
			Quantity: stripe.Int64(int64(candle.Quantity)),
		})
	}

	// 4. Create Stripe Checkout Session
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		Mode:               stripe.String("payment"),
		LineItems:          lineItems,
		SuccessURL:         stripe.String("http://localhost:3000/success?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:          stripe.String("http://localhost:3000/cancel"),
		ClientReferenceID:  stripe.String(orderID.String()), // link checkout session to your order
		CustomerEmail:      stripe.String(req.Email),
	}

	s, err := session.New(params)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to create Stripe session: " + err.Error(),
		})
		return
	}

	// 5. Return both order ID and checkout URL
	ctx.JSON(http.StatusCreated, gin.H{
		"order_id":     orderID,
		"checkout_url": s.URL,
	})
}

func (oh *OrderHandler) UpdatePayedOrder(ctx *gin.Context) {
	orderID := ctx.Param("order_id")
	sessionID := ctx.Query("session_id")

	if orderID == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Order ID is required",
		})
		return
	}

	err := oh.OrderService.UpdatePayedOrder(ctx, uuid.MustParse(orderID), sessionID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to update order status: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Order status updated to paid",
	})
}

func (oh *OrderHandler) HandleStripeWebhook(c *gin.Context) {
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)

	payload, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Println("Error reading body:", err)
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}

	endpointSecret := os.Getenv("STRIPE_ENDPOINT")
	sigHeader := c.GetHeader("Stripe-Signature")

	event, err := webhook.ConstructEvent(payload, sigHeader, endpointSecret)
	if err != nil {
		log.Println("Webhook signature verification failed:", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// Handle the event
	switch event.Type {
	case "checkout.session.completed":
		// Here you can:
		// 1. Retrieve client_reference_id (your order ID)
		// 2. Mark order as paid in your database
		log.Println("order_ID:", event.Data.Object["client_reference_id"].(string))
		log.Println("session_id:", event.Data.Object["id"].(string))
		log.Print("session_ID: ", event.ID)
		err := oh.OrderService.UpdatePayedOrder(context.Background(), uuid.MustParse(event.Data.Object["client_reference_id"].(string)), event.Data.Object["id"].(string))
		if err != nil {
			log.Println("Failed to update order status:", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	case "payment_intent.payment_failed":
		pi := event.Data.Object
		log.Println("Payment failed:", pi)
	default:
		log.Println("Unhandled event type:", event.Type)
	}

	c.Status(http.StatusOK)
}
