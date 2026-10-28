package gateway

import (
	"context"
	"fmt"

	"github.com/quickeats/services/payment/internal/domain"
	"github.com/quickeats/services/payment/internal/port"
)

// CardStrategy charges the customer's saved card. This demo has no real card network:
// each saved card holds its own balance, and a charge debits it immediately.
type CardStrategy struct {
	cards port.CardRepository
}

func NewCardStrategy(cards port.CardRepository) *CardStrategy {
	return &CardStrategy{cards: cards}
}

// Execute returns domain.ErrNoCard or domain.ErrInsufficientFunds when the charge is declined.
func (s *CardStrategy) Execute(ctx context.Context, payment *domain.Payment) (*domain.PaymentResult, error) {
	card, err := s.cards.GetByUserID(ctx, payment.CustomerID)
	if err != nil {
		return nil, err
	}
	if card == nil {
		return nil, domain.ErrNoCard
	}
	if err := s.cards.Debit(ctx, payment.CustomerID, payment.Amount); err != nil {
		return nil, err
	}
	return &domain.PaymentResult{
		GatewayRef: fmt.Sprintf("card:%s:%s", card.Brand, card.Last4),
		Completed:  true,
	}, nil
}
