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
	Currency            *string             `json:"currency"`
	Email               string              `json:"email" binding:"required,email"`
	FirstName           string              `json:"first_name"`
	LastName            string              `json:"last_name"`
	Phone               *string             `json:"phone"`
	Address             *domain.Address     `json:"address"`
	PickupPoint         *string             `json:"pickup_point"`
	Candles             []domain.CandleItem `json:"candles" binding:"required"`
	PromotionName       *string             `json:"promotion_name"`
	BillingAddressMatch bool                `json:"billing_address_match"`
	BillingCountry      *string             `json:"billing_country"`
	BillingCity         *string             `json:"billing_city"`
	BillingZip          *string             `json:"billing_zip"`
	BillingStreet       *string             `json:"billing_street"`
	BillingLine1        *string             `json:"billing_line1"`
}

type OrderHandler struct {
	OrderService     service.OrderService
	CandleService    service.CandleService
	MailService      service.MailService
	PromotionService service.PromotionService
}

func (oh *OrderHandler) OrderEndpoints(router *gin.Engine) {
	router.POST("orders", oh.CreateOrder)
	router.PUT("orders/:order_id", oh.UpdatePayedOrder)
	router.POST("orders/webhook", oh.HandleStripeWebhook)
}

func NewOrderHandler(orderService service.OrderService, candleService service.CandleService, mailService service.MailService, promotionService service.PromotionService) *OrderHandler {
	return &OrderHandler{
		orderService,
		candleService,
		mailService,
		promotionService,
	}
}

func (oh *OrderHandler) CreateOrder(ctx *gin.Context) {
	var req createOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("JSON binding error: %v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request: " + err.Error(),
		})
		return
	}
	log.Printf("CreateOrder request received: email=%s, currency=%v, candles=%d, promotion=%v", req.Email, req.Currency, len(req.Candles), req.PromotionName)

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

	// Validate billing address if billing_address_match is false
	if !req.BillingAddressMatch {
		if req.BillingCountry == nil || *req.BillingCountry == "" ||
			req.BillingCity == nil || *req.BillingCity == "" ||
			req.BillingZip == nil || *req.BillingZip == "" ||
			req.BillingStreet == nil || *req.BillingStreet == "" ||
			req.BillingLine1 == nil || *req.BillingLine1 == "" {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "Billing address fields are required when billing_address_match is false",
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
	log.Printf("Processing order with currency: %s", currency)

	// Get promotion if promotion_name is provided
	var promotion *domain.PromotionResponse
	if req.PromotionName != nil && *req.PromotionName != "" {
		log.Printf("Retrieving promotion: %s for email: %s", *req.PromotionName, req.Email)
		var err error
		promotion, err = oh.PromotionService.GetPromotionByName(ctx, *req.PromotionName, req.Email)
		if err != nil {
			log.Printf("Promotion retrieval failed: %v", err)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to retrieve promotion: " + err.Error(),
			})
			return
		}
		if promotion == nil {
			log.Printf("Promotion not found: %s", *req.PromotionName)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "Invalid promotion code",
			})
			return
		}

		if !promotion.Available {
			log.Printf("Promotion already used: %s by email: %s", *req.PromotionName, req.Email)
			ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{
				"message": "Promotion already used by this email",
			})
			return
		}

		log.Printf("Promotion applied: %s with %d%% discount", promotion.Promotion.Name, promotion.Promotion.Percentage)
	}

	for _, candle := range req.Candles {
		candleData, err := oh.CandleService.GetCandleByID(ctx, candle.ID.String())
		if err != nil || candleData == nil {
			log.Printf("Candle validation failed for ID %s: err=%v, candleData=%v", candle.ID.String(), err, candleData)
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
			log.Printf("Unsupported currency: %s", currency)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "Unsupported currency: " + currency,
			})
			return
		}
		if candle.Price != expectedPrice {
			log.Printf("Price mismatch for candle %s: expected=%d, got=%d", candle.ID.String(), expectedPrice, candle.Price)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "Candle price mismatch for candle ID: " + candle.ID.String(),
			})
			return
		}
	}

	// 1. Calculate total price, discounted price, and shipping
	var cartTotal int64 = 0
	for _, candle := range req.Candles {
		itemTotal := candle.Price * int64(candle.Quantity)
		cartTotal += itemTotal
		log.Printf("Candle: %s, Price: %d, Quantity: %d, Item Total: %d", candle.Name, candle.Price, candle.Quantity, itemTotal)
	}

	// Apply promotion discount to calculate discounted price
	var discountedPrice *int64
	var finalCartTotal int64 = cartTotal
	if promotion != nil {
		discountAmount := (cartTotal * int64(promotion.Promotion.Percentage)) / 100
		finalCartTotal = cartTotal - discountAmount
		discountedPrice = &finalCartTotal
		log.Printf("Applying %d%% discount to items total %d: -%d, Final: %d", promotion.Promotion.Percentage, cartTotal, discountAmount, finalCartTotal)
	}

	// Calculate shipping cost
	var freeShippingThreshold int64
	var homeDeliveryShipping int64
	var pickupPointShipping int64

	switch currency {
	case "huf":
		freeShippingThreshold = 15000
		homeDeliveryShipping = 2100
		pickupPointShipping = 1400
	case "eur":
		freeShippingThreshold = 40
		homeDeliveryShipping = 6
		pickupPointShipping = 4
	case "czk":
		freeShippingThreshold = 1000
		homeDeliveryShipping = 149
		pickupPointShipping = 89
	default:
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Unsupported currency: " + currency,
		})
		return
	}

	var shippingPrice int64 = 0
	if finalCartTotal < freeShippingThreshold {
		if req.Address != nil {
			shippingPrice = homeDeliveryShipping
		} else {
			shippingPrice = pickupPointShipping
		}
		log.Printf("Adding shipping: %d %s", shippingPrice, currency)
	} else {
		log.Printf("Free shipping applies - cart total %d meets threshold %d", finalCartTotal, freeShippingThreshold)
	}

	// 2. Create order in DB with prices
	var promotionID *string
	if promotion != nil {
		promotionID = &promotion.Promotion.ID
	}
	orderID, err := oh.OrderService.CreateOrder(ctx, req.Email, req.FirstName, req.LastName, req.Phone, promotionID, cartTotal, discountedPrice, shippingPrice, req.BillingAddressMatch, req.BillingCountry, req.BillingCity, req.BillingZip, req.BillingStreet, req.BillingLine1)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to create order: " + err.Error(),
		})
		return
	}

	// 3. Add items to order
	for _, candle := range req.Candles {
		err = oh.OrderService.AddCandlesToOrder(ctx, orderID, candle.ID, candle.Quantity)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to add candles to order: " + err.Error(),
			})
			return
		}
	}

	// 4. Add address to order
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
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to add address to order: " + err.Error(),
			})
			return
		}
	}

	// 5. Convert candles into Stripe line items
	stripe.Key = os.Getenv("STRIPE_SECRET")
	var lineItems []*stripe.CheckoutSessionLineItemParams

	// Track number of candle items (before adding shipping)
	var numCandleItems int

	for _, candle := range req.Candles {
		lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String(currency),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name: stripe.String(candle.Name),
				},
				UnitAmount: stripe.Int64(candle.Price * 100),
			},
			Quantity: stripe.Int64(int64(candle.Quantity)),
		})
		numCandleItems++
	}

	// Apply promotion discount proportionally to candle items only
	if promotion != nil {
		for i := 0; i < numCandleItems; i++ {
			originalAmount := lineItems[i].PriceData.UnitAmount
			discountedAmount := (*originalAmount * (100 - int64(promotion.Promotion.Percentage))) / 100
			lineItems[i].PriceData.UnitAmount = stripe.Int64(discountedAmount)
		}
	}

	// Add shipping as a line item if applicable
	if shippingPrice > 0 {
		var shippingName string
		if req.Address != nil {
			shippingName = "Home Delivery"
		} else {
			shippingName = "Pickup Point Delivery"
		}

		lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String(currency),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name: stripe.String(shippingName),
				},
				UnitAmount: stripe.Int64(shippingPrice * 100),
			},
			Quantity: stripe.Int64(1),
		})
	}

	// 6. Create Stripe Checkout Session
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

	// 7. Return both order ID and checkout URL
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

	// Get order details with candles for email notification
	orderWithCandles, err := oh.OrderService.GetOrderWithCandles(ctx, uuid.MustParse(orderID))
	if err != nil {
		log.Println("Failed to get order details for notification: " + err.Error())
	} else {
		err = oh.MailService.SendOrderNotification(orderWithCandles)
		if err != nil {
			log.Println("Failed to send order notification email: " + err.Error())
		}
		err = oh.MailService.SendOrderConfirmation(orderWithCandles)
		if err != nil {
			log.Println("Failed to send order confirmation email: " + err.Error())
		}
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
		orderID := uuid.MustParse(event.Data.Object["client_reference_id"].(string))
		err := oh.OrderService.UpdatePayedOrder(context.Background(), orderID, event.Data.Object["id"].(string))
		if err != nil {
			log.Println("Failed to update order status:", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Send order notification with full details
		orderWithCandles, err := oh.OrderService.GetOrderWithCandles(context.Background(), orderID)
		if err != nil {
			log.Println("Failed to get order details for notification: " + err.Error())
		} else {
			// Save promotion usage if promotion was applied
			if orderWithCandles.Order.PromotionID != nil && *orderWithCandles.Order.PromotionID != "" {
				err = oh.PromotionService.SavePromotionUsage(context.Background(), *orderWithCandles.Order.PromotionID, orderWithCandles.Order.Email)
				if err != nil {
					log.Println("Failed to save promotion usage: " + err.Error())
				} else {
					log.Printf("Promotion usage saved: promotion_id=%s, email=%s", *orderWithCandles.Order.PromotionID, orderWithCandles.Order.Email)
				}
			}

			err = oh.MailService.SendOrderNotification(orderWithCandles)
			if err != nil {
				log.Println("Failed to send order notification email: " + err.Error())
			}
		}
	case "payment_intent.payment_failed":
		pi := event.Data.Object
		log.Println("Payment failed:", pi)
	default:
		log.Println("Unhandled event type:", event.Type)
	}

	c.Status(http.StatusOK)
}
