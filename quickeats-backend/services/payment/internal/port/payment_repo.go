package port

import (
	"context"
	"github.com/quickeats/services/payment/internal/domain"
)

type PaymentRepository interface {
	Create(ctx context.Context, p *domain.Payment) error
	GetByID(ctx context.Context, id string) (*domain.Payment, error)
	GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error)
	UpdateStatus(ctx context.Context, id string, status domain.PaymentStatus, gatewayRef string) error
	// RefundToCard marks a COMPLETED payment REFUNDED and returns its amount to the
	// customer's card in one transaction. It reports false, changing nothing, if the
	// payment is not COMPLETED (never charged, or already refunded).
	RefundToCard(ctx context.Context, paymentID, userID string) (bool, error)
}

type LedgerRepository interface {
	CreateEntries(ctx context.Context, entries []*domain.LedgerEntry) error
	GetByPaymentID(ctx context.Context, paymentID string) ([]*domain.LedgerEntry, error)
}

type CardRepository interface {
	// Save stores the customer's card, replacing any card they had before.
	Save(ctx context.Context, card *domain.Card) error
	GetByUserID(ctx context.Context, userID string) (*domain.Card, error)
	Delete(ctx context.Context, userID string) error
	// Debit takes amount from the customer's card. It returns domain.ErrNoCard or
	// domain.ErrInsufficientFunds without changing anything when it can't.
	Debit(ctx context.Context, userID string, amount int64) error
}
