package domain

import "context"

type PaymentResult struct {
	GatewayRef         string `json:"gatewayRef,omitempty"`
	GatewayRedirectURL string `json:"gatewayRedirectUrl,omitempty"`
	Completed          bool   `json:"completed"` // true for wallet (instant), false for card (async)
}

// PaymentStrategy defines how a payment method processes a charge.
type PaymentStrategy interface {
	Execute(ctx context.Context, payment *Payment) (*PaymentResult, error)
}
