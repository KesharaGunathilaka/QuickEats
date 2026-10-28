package port

import "context"

type CreatePaymentIntentRequest struct {
	OrderID    string `json:"orderId"`
	CustomerID string `json:"customerId"`
	ShopID     string `json:"shopId"`
	Amount     int64  `json:"amount"`
	Currency   string `json:"currency"`
	Method     string `json:"method"`
}

type PaymentIntentResponse struct {
	PaymentID          string `json:"paymentId"`
	Status             string `json:"status"`
	GatewayRedirectURL string `json:"gatewayRedirectUrl,omitempty"`
}

// CardCheck says whether the customer's card can pay an amount (the balance itself
// stays inside the payment service).
type CardCheck struct {
	HasCard    bool `json:"hasCard"`
	Sufficient bool `json:"sufficient"`
}

// Reasons the payment service gives for a declined payment; stored as the order's
// cancel_reason so the app can explain what happened.
const (
	DeclineNoCard            = "no_card"
	DeclineInsufficientFunds = "insufficient_funds"
	DeclineOther             = "payment_failed"
)

// PaymentDeclinedError is returned by CreatePaymentIntent when the card was refused.
type PaymentDeclinedError struct {
	Reason string // one of the Decline* codes
}

func (e *PaymentDeclinedError) Error() string { return "payment declined: " + e.Reason }

type PaymentClient interface {
	CreatePaymentIntent(ctx context.Context, req CreatePaymentIntentRequest) (*PaymentIntentResponse, error)
	// RefundOrder returns the order's payment to the customer's card. It is safe to
	// call for an order that was never charged or already refunded.
	RefundOrder(ctx context.Context, orderID, customerID string) error
	CheckCard(ctx context.Context, customerID string, amount int64) (CardCheck, error)
}
