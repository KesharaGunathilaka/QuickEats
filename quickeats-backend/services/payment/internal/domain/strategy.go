package domain

import "context"

type PaymentResult struct {
	GatewayRef         string `json:"gatewayRef,omitempty"`
	GatewayRedirectURL string `json:"gatewayRedirectUrl,omitempty"`
	Completed          bool   `json:"completed"` // true when charged immediately, false if a gateway confirms later
}

// PaymentStrategy defines how a payment method processes a charge.
type PaymentStrategy interface {
	Execute(ctx context.Context, payment *Payment) (*PaymentResult, error)
}
