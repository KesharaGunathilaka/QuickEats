package app

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/quickeats/pkg/events"
	"github.com/quickeats/pkg/outbox"
	"github.com/quickeats/services/payment/internal/domain"
	"github.com/quickeats/services/payment/internal/port"
)

type CreatePaymentUseCase struct {
	db          *pgxpool.Pool
	paymentRepo port.PaymentRepository
	ledgerRepo  port.LedgerRepository
	walletRepo  port.WalletRepository
	strategies  map[string]domain.PaymentStrategy
}

func NewCreatePaymentUseCase(db *pgxpool.Pool, paymentRepo port.PaymentRepository, ledgerRepo port.LedgerRepository, walletRepo port.WalletRepository, strategies map[string]domain.PaymentStrategy) *CreatePaymentUseCase {
	return &CreatePaymentUseCase{db: db, paymentRepo: paymentRepo, ledgerRepo: ledgerRepo, walletRepo: walletRepo, strategies: strategies}
}

type CreateIntentRequest struct {
	OrderID    string `json:"orderId"`
	CustomerID string `json:"customerId"`
	Amount     int64  `json:"amount"`
	Currency   string `json:"currency"`
	Method     string `json:"method"`
}

type IntentResponse struct {
	PaymentID          string `json:"paymentId"`
	Status             string `json:"status"`
	GatewayRedirectURL string `json:"gatewayRedirectUrl,omitempty"`
}

func (uc *CreatePaymentUseCase) Execute(ctx context.Context, req CreateIntentRequest) (*IntentResponse, error) {
	// Check for duplicate
	existing, _ := uc.paymentRepo.GetByOrderID(ctx, req.OrderID)
	if existing != nil {
		return &IntentResponse{PaymentID: existing.ID, Status: string(existing.Status)}, nil
	}

	payment := domain.NewPayment(req.OrderID, req.Method, req.Amount, req.Currency)

	// Resolve strategy
	strategy, ok := uc.strategies[req.Method]
	if !ok {
		return nil, fmt.Errorf("unsupported payment method: %s", req.Method)
	}

	// Create payment record
	if err := uc.paymentRepo.Create(ctx, payment); err != nil {
		return nil, err
	}

	// Execute payment strategy
	result, err := strategy.Execute(ctx, payment)
	if err != nil {
		// Publish payment.failed
		uc.publishPaymentFailed(ctx, payment, err.Error())
		return nil, err
	}

	resp := &IntentResponse{PaymentID: payment.ID, Status: string(payment.Status)}

	if result.Completed {
		// Wallet payment: complete immediately
		uc.completePayment(ctx, payment, result.GatewayRef, req.CustomerID)
		resp.Status = string(domain.PaymentCompleted)
	} else {
		// Card: return redirect URL, wait for webhook
		resp.GatewayRedirectURL = result.GatewayRedirectURL
	}

	return resp, nil
}

func (uc *CreatePaymentUseCase) completePayment(ctx context.Context, payment *domain.Payment, gatewayRef, customerID string) {
	uc.paymentRepo.UpdateStatus(ctx, payment.ID, domain.PaymentCompleted, gatewayRef)

	// Double-entry ledger
	entries := []*domain.LedgerEntry{
		domain.NewLedgerEntry(payment.ID, fmt.Sprintf("customer:%s", customerID), domain.Debit, payment.Amount, "order payment"),
		domain.NewLedgerEntry(payment.ID, fmt.Sprintf("shop:%s", payment.OrderID), domain.Credit, payment.Amount*95/100, "order revenue (95%)"),
		domain.NewLedgerEntry(payment.ID, "platform:revenue", domain.Credit, payment.Amount*5/100, "platform commission (5%)"),
	}
	uc.ledgerRepo.CreateEntries(ctx, entries)

	// Publish completed event via outbox
	tx, err := uc.db.Begin(ctx)
	if err != nil { return }
	defer tx.Rollback(ctx)

	event := events.NewEvent("payment.payment.completed", "payment", payment.ID, map[string]interface{}{
		"payment_id": payment.ID, "order_id": payment.OrderID,
		"method": payment.Method, "amount": payment.Amount,
		"currency": payment.Currency, "gateway_ref": gatewayRef,
	})
	outbox.WriteToOutbox(ctx, tx, event)
	tx.Commit(ctx)
}

func (uc *CreatePaymentUseCase) publishPaymentFailed(ctx context.Context, payment *domain.Payment, reason string) {
	uc.paymentRepo.UpdateStatus(ctx, payment.ID, domain.PaymentFailed, "")
	tx, _ := uc.db.Begin(ctx)
	defer tx.Rollback(ctx)
	event := events.NewEvent("payment.payment.failed", "payment", payment.ID, map[string]interface{}{
		"payment_id": payment.ID, "order_id": payment.OrderID, "reason": reason,
	})
	outbox.WriteToOutbox(ctx, tx, event)
	tx.Commit(ctx)
}
