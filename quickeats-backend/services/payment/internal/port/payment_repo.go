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
}

type LedgerRepository interface {
	CreateEntries(ctx context.Context, entries []*domain.LedgerEntry) error
	GetByPaymentID(ctx context.Context, paymentID string) ([]*domain.LedgerEntry, error)
}

type WalletRepository interface {
	Create(ctx context.Context, w *domain.Wallet) error
	GetByUserID(ctx context.Context, userID string) (*domain.Wallet, error)
	Debit(ctx context.Context, userID string, amount int64) error
	Credit(ctx context.Context, userID string, amount int64) error
}
