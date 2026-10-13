package gateway

import (
	"context"
	"fmt"

	"github.com/quickeats/services/payment/internal/domain"
	"github.com/quickeats/services/payment/internal/port"
)

type WalletStrategy struct {
	walletRepo port.WalletRepository
}

func NewWalletStrategy(walletRepo port.WalletRepository) *WalletStrategy {
	return &WalletStrategy{walletRepo: walletRepo}
}

func (s *WalletStrategy) Execute(ctx context.Context, payment *domain.Payment) (*domain.PaymentResult, error) {
	// Find wallet by extracting customer ID from context or payment metadata
	// For simplicity, the order's customer owns the wallet
	wallet, err := s.walletRepo.GetByUserID(ctx, "") // Customer ID should be passed via payment metadata
	if err != nil || wallet == nil {
		return nil, domain.ErrWalletNotFound
	}
	if wallet.Balance < payment.Amount {
		return nil, domain.ErrWalletInsufficientFunds
	}

	if err := s.walletRepo.Debit(ctx, wallet.UserID, payment.Amount); err != nil {
		return nil, fmt.Errorf("wallet debit: %w", err)
	}

	return &domain.PaymentResult{
		GatewayRef: fmt.Sprintf("wallet:%s", wallet.ID),
		Completed:  true,
	}, nil
}
