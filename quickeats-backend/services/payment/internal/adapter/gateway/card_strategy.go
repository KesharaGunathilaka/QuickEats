package gateway

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/quickeats/services/payment/internal/domain"
)

type CardStrategy struct {
	merchantID     string
	merchantSecret string
	sandbox        bool
}

func NewCardStrategy(merchantID, merchantSecret string, sandbox bool) *CardStrategy {
	return &CardStrategy{merchantID: merchantID, merchantSecret: merchantSecret, sandbox: sandbox}
}

func (s *CardStrategy) Execute(ctx context.Context, payment *domain.Payment) (*domain.PaymentResult, error) {
	// In production: generate a PayHere/Stripe checkout session URL
	// For now: return a simulated redirect URL
	sessionID := uuid.New().String()

	// Local dev: no gateway credentials configured, so simulate an instant successful charge
	if s.merchantID == "" {
		return &domain.PaymentResult{GatewayRef: fmt.Sprintf("simulated:%s", sessionID), Completed: true}, nil
	}

	baseURL := "https://sandbox.payhere.lk"
	if !s.sandbox {
		baseURL = "https://www.payhere.lk"
	}

	return &domain.PaymentResult{
		GatewayRef:         sessionID,
		GatewayRedirectURL: fmt.Sprintf("%s/pay/checkout/%s", baseURL, sessionID),
		Completed:          false, // Awaits webhook callback
	}, nil
}
