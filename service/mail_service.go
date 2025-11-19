package service

import (
	"fmt"
	"strings"

	"elindor/domain"

	"gopkg.in/gomail.v2"
)

type MailService interface {
	SendOrderNotification(orderWithCandles *domain.OrderWithCandles) error
	SendOrderConfirmation(orderWithCandles *domain.OrderWithCandles) error
}

type mailService struct {
	dialer gomail.Dialer
}

func NewMailService(dialer gomail.Dialer) MailService {
	return &mailService{
		dialer: dialer,
	}
}

func (ms *mailService) SendOrderNotification(orderWithCandles *domain.OrderWithCandles) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "hello@elindorcandle.com")
	m.SetHeader("To", "sooskristof123@gmail.com", "vilmossoos26@gmail.com", "hello@elindorcandle.com")
	m.SetHeader("Subject", "New Order Received - ELINDOR")

	// Build detailed order information
	var totalAmount float64
	var candlesList strings.Builder

	candlesList.WriteString("Order Details:\n")
	candlesList.WriteString("=============\n\n")
	candlesList.WriteString(fmt.Sprintf("Order ID: %s\n", orderWithCandles.Order.ID.String()))
	candlesList.WriteString(fmt.Sprintf("Customer Name: %s %s\n", orderWithCandles.Order.FirstName, orderWithCandles.Order.LastName))
	candlesList.WriteString(fmt.Sprintf("Customer Email: %s\n", orderWithCandles.Order.Email))
	candlesList.WriteString(fmt.Sprintf("Status: %s\n\n", orderWithCandles.Order.Status))

	// Add delivery information
	if orderWithCandles.Order.IsHomeDelivery {
		candlesList.WriteString("Delivery Type: Home Delivery\n")
		if orderWithCandles.Order.Country != nil {
			candlesList.WriteString(fmt.Sprintf("Address: %s", *orderWithCandles.Order.Street))
			if orderWithCandles.Order.Line1 != nil && *orderWithCandles.Order.Line1 != "" {
				candlesList.WriteString(fmt.Sprintf(", %s", *orderWithCandles.Order.Line1))
			}
			candlesList.WriteString(fmt.Sprintf("\n%s %s, %s\n", *orderWithCandles.Order.Zipcode, *orderWithCandles.Order.City, *orderWithCandles.Order.Country))
		}
	} else if orderWithCandles.Order.PickupPoint != nil {
		candlesList.WriteString(fmt.Sprintf("Delivery Type: Pickup Point - %s\n", *orderWithCandles.Order.PickupPoint))
	}

	candlesList.WriteString("\nOrdered Items:\n")
	candlesList.WriteString("-------------\n")

	for _, orderCandle := range orderWithCandles.Candles {
		candlesList.WriteString(fmt.Sprintf("- %s (%s)\n", orderCandle.Candle.NameHU, orderCandle.Candle.NameEN))
		candlesList.WriteString(fmt.Sprintf("  Quantity: %d\n", orderCandle.Quantity))
		candlesList.WriteString(fmt.Sprintf("  Price (HUF): %.0f Ft each\n", orderCandle.Candle.PriceHUF))
		itemTotal := orderCandle.Candle.PriceHUF * float64(orderCandle.Quantity)
		candlesList.WriteString(fmt.Sprintf("  Subtotal: %.0f Ft\n\n", itemTotal))
		totalAmount += itemTotal
	}

	candlesList.WriteString(fmt.Sprintf("TOTAL AMOUNT: %.0f Ft\n", totalAmount))

	if orderWithCandles.Order.SessionID != nil {
		candlesList.WriteString(fmt.Sprintf("Stripe Session ID: %s\n", *orderWithCandles.Order.SessionID))
	}

	m.SetBody("text/plain", candlesList.String())

	if err := ms.dialer.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

func (ms *mailService) SendOrderConfirmation(orderWithCandles *domain.OrderWithCandles) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "hello@elindorcandle.com")
	m.SetHeader("To", orderWithCandles.Order.Email)
	m.SetHeader("Subject", "Thank you for your order - ELINDOR")

	htmlTemplate := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Order Confirmation</title>
    <style>
        body {
            font-family: 'Georgia', serif;
            background-color: #fcf9ee;
            margin: 0;
            padding: 40px;
            color: #212121;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            background-color: #fcf9ee;
            padding: 40px;
            box-shadow: 0 0 20px rgba(0,0,0,0.1);
        }
        .header {
            text-align: center;
            margin-bottom: 40px;
        }
        .brand {
            font-size: 32px;
            font-weight: normal;
            letter-spacing: 8px;
            color: #212121;
            margin-bottom: 40px;
        }
        .thank-you {
            font-size: 20px;
            margin-bottom: 10px;
            color: #212121;
        }
        .order-number {
            font-size: 14px;
            color: #888;
            margin-bottom: 40px;
        }
        .addresses {
            display: flex;
            justify-content: space-between;
            margin-bottom: 40px;
        }
        .address-section {
            width: 48%;
        }
        .address-title {
            font-size: 14px;
            color: #888;
            margin-bottom: 15px;
        }
        .address {
            font-size: 14px;
            line-height: 1.6;
            color: #333;
        }
        .products-table {
            width: 100%;
            border-collapse: collapse;
            margin-bottom: 30px;
        }
        .products-table th {
            text-align: left;
            font-size: 14px;
            color: #888;
            border-bottom: 1px solid #ddd;
            padding: 15px 0;
        }
        .products-table td {
            padding: 20px 0;
            border-bottom: 1px solid #eee;
            vertical-align: top;
        }
        .product-image {
            width: 60px;
            height: 60px;
            border-radius: 4px;
        }
        .product-name {
            font-size: 14px;
            color: #333;
            margin-bottom: 5px;
        }
        .product-details {
            font-size: 12px;
            color: #666;
        }
        .qty, .price {
            text-align: center;
            font-size: 14px;
            color: #333;
        }
        .totals {
            padding-top: 20px;
        }
        .total-row {
            display: flex;
            justify-content: space-between;
            margin-bottom: 10px;
            font-size: 14px;
        }
        .total-final {
            font-weight: bold;
            font-size: 16px;
            border-top: 1px solid #ddd;
            padding-top: 15px;
            margin-top: 15px;
        }
        .help-section {
            text-align: center;
            margin-top: 50px;
        }
        .help-title {
            font-size: 18px;
            margin-bottom: 15px;
            color: #333;
        }
        .help-text {
            font-size: 14px;
            color: #666;
            margin-bottom: 25px;
        }
        .contact-button {
            background-color: #333;
            color: white;
            padding: 12px 30px;
            text-decoration: none;
            font-size: 14px;
            display: inline-block;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="brand">ELINDOR</div>
            <div class="thank-you">Thank you for your order</div>
            <div class="order-number">Order # {{ORDER_ID}}</div>
        </div>

        <div class="addresses">
            <div class="address-section">
                <div class="address-title">Shipping Address</div>
                <div class="address">
                    {{SHIPPING_ADDRESS}}
                </div>
            </div>
            <div class="address-section">
                <div class="address-title">Billing Address</div>
                <div class="address">
                    {{BILLING_ADDRESS}}
                </div>
            </div>
        </div>

        <table class="products-table">
            <thead>
                <tr>
                    <th style="width: 60%;">Product</th>
                    <th style="width: 20%; text-align: center;">Qty</th>
                    <th style="width: 20%; text-align: center;">Price</th>
                </tr>
            </thead>
            <tbody>
                {{PRODUCTS_LIST}}
            </tbody>
        </table>

        <div class="totals">
            <table style="width: 100%; border: none;">
                <tr>
                    <td style="text-align: left; padding: 5px 0; font-size: 14px;">Items ({{TOTAL_ITEMS}})</td>
                    <td style="text-align: right; padding: 5px 0; font-size: 14px;">{{TOTAL_AMOUNT}}</td>
                </tr>
                <tr>
                    <td style="text-align: left; padding: 5px 0; font-size: 14px;">Shipping</td>
                    <td style="text-align: right; padding: 5px 0; font-size: 14px;">Free</td>
                </tr>
                <tr style="border-top: 1px solid #ddd;">
                    <td style="text-align: left; padding: 15px 0 5px 0; font-size: 16px; font-weight: bold;">TOTAL</td>
                    <td style="text-align: right; padding: 15px 0 5px 0; font-size: 16px; font-weight: bold;">{{TOTAL_AMOUNT}}</td>
                </tr>
            </table>
        </div>

        <div class="help-section">
            <div class="help-title">Do you need help?</div>
            <div class="help-text">If you have any questions, feel free to send us a message.</div>
            <a href="mailto:hello@elindorcandle.com" class="contact-button">Contact Us</a>
        </div>
    </div>
</body>
</html>`

	// Replace placeholders with actual data
	body := htmlTemplate
	body = strings.Replace(body, "{{ORDER_ID}}", orderWithCandles.Order.ID.String(), 1)
	body = strings.Replace(body, "{{CUSTOMER_EMAIL}}", orderWithCandles.Order.Email, -1)

	// Calculate totals
	var totalAmount float64
	var totalItems int
	for _, orderCandle := range orderWithCandles.Candles {
		totalAmount += orderCandle.Candle.PriceHUF * float64(orderCandle.Quantity)
		totalItems += orderCandle.Quantity
	}

	body = strings.Replace(body, "{{TOTAL_ITEMS}}", fmt.Sprintf("%d", totalItems), 1)
	body = strings.Replace(body, "{{TOTAL_AMOUNT}}", fmt.Sprintf("%.0f Ft", totalAmount), -1)

	// Generate address information
	customerName := fmt.Sprintf(`%s %s`, orderWithCandles.Order.FirstName, orderWithCandles.Order.LastName)
	var shippingAddress, billingAddress string
	if orderWithCandles.Order.IsHomeDelivery {
		if orderWithCandles.Order.Street != nil {
			shippingAddress = fmt.Sprintf(`%s<br>%s`, customerName, *orderWithCandles.Order.Street)
			if orderWithCandles.Order.Line1 != nil && *orderWithCandles.Order.Line1 != "" {
				shippingAddress += fmt.Sprintf(`, %s`, *orderWithCandles.Order.Line1)
			}
			shippingAddress += "<br>"
			if orderWithCandles.Order.Zipcode != nil && orderWithCandles.Order.City != nil {
				shippingAddress += fmt.Sprintf(`%s %s<br>`, *orderWithCandles.Order.Zipcode, *orderWithCandles.Order.City)
			}
			if orderWithCandles.Order.Country != nil {
				shippingAddress += fmt.Sprintf(`%s<br><br>`, *orderWithCandles.Order.Country)
			}
			shippingAddress += fmt.Sprintf(`%s<br>+36 70 123 4567`, orderWithCandles.Order.Email)
			billingAddress = shippingAddress // Same for now
		} else {
			shippingAddress = fmt.Sprintf(`%s<br>Home Delivery Address<br><br>%s<br>+36 70 123 4567`, customerName, orderWithCandles.Order.Email)
			billingAddress = shippingAddress
		}
	} else if orderWithCandles.Order.PickupPoint != nil {
		shippingAddress = fmt.Sprintf(`%s<br>Pickup Point: %s<br><br>%s<br>+36 70 123 4567`, customerName, *orderWithCandles.Order.PickupPoint, orderWithCandles.Order.Email)
		billingAddress = shippingAddress
	} else {
		shippingAddress = fmt.Sprintf(`%s<br>Address not provided<br><br>%s<br>+36 70 123 4567`, customerName, orderWithCandles.Order.Email)
		billingAddress = shippingAddress
	}

	body = strings.Replace(body, "{{SHIPPING_ADDRESS}}", shippingAddress, 1)
	body = strings.Replace(body, "{{BILLING_ADDRESS}}", billingAddress, 1)

	// Generate products list
	var productsHTML strings.Builder
	for _, orderCandle := range orderWithCandles.Candles {
		productHTML := fmt.Sprintf(`
                <tr>
                    <td>
                        <div style="display: flex; align-items: center;">
                            <img src="%s" alt="%s" class="product-image" style="margin-right: 15px;">
                            <div>
                                <div class="product-name">%s</div>
                                <div class="product-details">200g</div>
                            </div>
                        </div>
                    </td>
                    <td class="qty">%d</td>
                    <td class="price">%.0f Ft</td>
                </tr>`,
			orderCandle.Candle.ImageURL,
			orderCandle.Candle.NameHU,
			orderCandle.Candle.NameHU,
			orderCandle.Quantity,
			orderCandle.Candle.PriceHUF*float64(orderCandle.Quantity))
		productsHTML.WriteString(productHTML)
	}

	body = strings.Replace(body, "{{PRODUCTS_LIST}}", productsHTML.String(), 1)

	m.SetBody("text/html", body)

	if err := ms.dialer.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
