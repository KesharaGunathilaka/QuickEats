package app

import (
	"context"
	"time"

	"github.com/quickeats/services/payment/internal/domain"
	"github.com/quickeats/services/payment/internal/port"
)

// CardUseCase manages the customer's saved card.
type CardUseCase struct {
	cards port.CardRepository
	now   func() time.Time
}

func NewCardUseCase(cards port.CardRepository) *CardUseCase {
	return &CardUseCase{cards: cards, now: time.Now}
}

type AddCardRequest struct {
	Number     string `json:"number"`
	HolderName string `json:"holderName"`
	ExpMonth   int    `json:"expMonth"`
	ExpYear    int    `json:"expYear"`
	CVV        string `json:"cvv"`
}

// AddCard validates and saves a card, replacing the customer's current one. The new
// card starts with the demo balance (domain.DemoCardBalance).
func (uc *CardUseCase) AddCard(ctx context.Context, userID string, req AddCardRequest) (*domain.Card, error) {
	card, err := domain.NewCard(userID, req.Number, req.HolderName, req.ExpMonth, req.ExpYear, req.CVV, uc.now())
	if err != nil {
		return nil, err
	}
	if err := uc.cards.Save(ctx, card); err != nil {
		return nil, err
	}
	return card, nil
}

// GetCard returns the customer's card, or nil if they have none.
func (uc *CardUseCase) GetCard(ctx context.Context, userID string) (*domain.Card, error) {
	return uc.cards.GetByUserID(ctx, userID)
}

func (uc *CardUseCase) RemoveCard(ctx context.Context, userID string) error {
	return uc.cards.Delete(ctx, userID)
}

// PayCheck answers whether the customer can pay an amount, without revealing the balance.
type PayCheck struct {
	HasCard    bool `json:"hasCard"`
	Sufficient bool `json:"sufficient"`
}

func (uc *CardUseCase) CanPay(ctx context.Context, userID string, amount int64) (PayCheck, error) {
	card, err := uc.cards.GetByUserID(ctx, userID)
	if err != nil {
		return PayCheck{}, err
	}
	if card == nil {
		return PayCheck{}, nil
	}
	return PayCheck{HasCard: true, Sufficient: card.Balance >= amount}, nil
}
