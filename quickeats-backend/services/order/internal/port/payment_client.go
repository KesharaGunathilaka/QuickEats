package port

import "context"

type CreatePaymentIntentRequest struct {
	OrderID    string `json:"orderId"`
	CustomerID string `json:"customerId"`
	Amount     int64  `json:"amount"`
	Currency   string `json:"currency"`
	Method     string `json:"method"`
}

type PaymentIntentResponse struct {
	PaymentID          string `json:"paymentId"`
	Status             string `json:"status"`
	GatewayRedirectURL string `json:"gatewayRedirectUrl,omitempty"`
}

type PaymentClient interface {
	CreatePaymentIntent(ctx context.Context, req CreatePaymentIntentRequest) (*PaymentIntentResponse, error)
	Refund(ctx context.Context, paymentID, orderID string, amount int64, reason string) error
}
