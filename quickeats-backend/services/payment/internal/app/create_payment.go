package app

import (
	"context"
	"fmt"
	"log/slog"

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
	strategies  map[string]domain.PaymentStrategy
}

func NewCreatePaymentUseCase(db *pgxpool.Pool, paymentRepo port.PaymentRepository, ledgerRepo port.LedgerRepository, strategies map[string]domain.PaymentStrategy) *CreatePaymentUseCase {
	return &CreatePaymentUseCase{db: db, paymentRepo: paymentRepo, ledgerRepo: ledgerRepo, strategies: strategies}
}

type CreateIntentRequest struct {
	OrderID    string `json:"orderId"`
	CustomerID string `json:"customerId"`
	ShopID     string `json:"shopId"`
	Amount     int64  `json:"amount"`
	Currency   string `json:"currency"`
	Method     string `json:"method"`
}

type IntentResponse struct {
	PaymentID          string `json:"paymentId"`
	Status             string `json:"status"`
	GatewayRedirectURL string `json:"gatewayRedirectUrl,omitempty"`
}

// DeclinedError is returned when the charge was refused; Reason is a domain.Failure* code.
type DeclinedError struct {
	Reason string
	Err    error
}

func (e *DeclinedError) Error() string { return e.Err.Error() }
func (e *DeclinedError) Unwrap() error { return e.Err }

func (uc *CreatePaymentUseCase) Execute(ctx context.Context, req CreateIntentRequest) (*IntentResponse, error) {
	// One payment per order: a repeated request reports the first attempt's outcome.
	existing, err := uc.paymentRepo.GetByOrderID(ctx, req.OrderID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if existing.Status == domain.PaymentFailed {
			// gateway_ref holds the failure reason for declined payments
			return nil, &DeclinedError{Reason: existing.GatewayRef, Err: fmt.Errorf("payment for order %s was declined", req.OrderID)}
		}
		return &IntentResponse{PaymentID: existing.ID, Status: string(existing.Status)}, nil
	}

	if req.Method == "" {
		req.Method = "card"
	}
	strategy, ok := uc.strategies[req.Method]
	if !ok {
		return nil, fmt.Errorf("unsupported payment method: %s", req.Method)
	}

	payment := domain.NewPayment(req.OrderID, req.Method, req.Amount, req.Currency)
	payment.CustomerID = req.CustomerID
	payment.ShopID = req.ShopID
	if err := uc.paymentRepo.Create(ctx, payment); err != nil {
		return nil, err
	}

	result, err := strategy.Execute(ctx, payment)
	if err != nil {
		reason := domain.FailureReason(err)
		uc.publishPaymentFailed(ctx, payment, reason)
		return nil, &DeclinedError{Reason: reason, Err: err}
	}

	resp := &IntentResponse{PaymentID: payment.ID, Status: string(payment.Status)}
	if result.Completed {
		uc.completePayment(ctx, payment, result.GatewayRef, req.CustomerID)
		resp.Status = string(domain.PaymentCompleted)
	} else {
		resp.GatewayRedirectURL = result.GatewayRedirectURL
	}
	return resp, nil
}

// RefundOrder returns a completed payment for the order to the customer's card.
// Calling it again, or for an order that was never charged, does nothing.
func (uc *CreatePaymentUseCase) RefundOrder(ctx context.Context, orderID, customerID string) (bool, error) {
	payment, err := uc.paymentRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return false, err
	}
	if payment == nil {
		return false, nil
	}
	refunded, err := uc.paymentRepo.RefundToCard(ctx, payment.ID, customerID)
	if err != nil || !refunded {
		return false, err
	}

	if err := uc.ledgerRepo.CreateEntries(ctx, refundEntries(payment, customerID)); err != nil {
		slog.Error("refund ledger entries failed", "payment_id", payment.ID, "error", err)
	}
	slog.Info("payment refunded to card", "order_id", orderID, "amount", payment.Amount)
	return true, nil
}

func (uc *CreatePaymentUseCase) completePayment(ctx context.Context, payment *domain.Payment, gatewayRef, customerID string) {
	uc.paymentRepo.UpdateStatus(ctx, payment.ID, domain.PaymentCompleted, gatewayRef)

	uc.ledgerRepo.CreateEntries(ctx, paymentEntries(payment, customerID))

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

// publishPaymentFailed records the decline (the reason goes in gateway_ref) and tells
// the order service, which cancels the order with that reason.
func (uc *CreatePaymentUseCase) publishPaymentFailed(ctx context.Context, payment *domain.Payment, reason string) {
	if err := uc.paymentRepo.UpdateStatus(ctx, payment.ID, domain.PaymentFailed, reason); err != nil {
		slog.Error("mark payment failed", "payment_id", payment.ID, "error", err)
	}
	tx, err := uc.db.Begin(ctx)
	if err != nil {
		slog.Error("publish payment failed: begin tx", "payment_id", payment.ID, "error", err)
		return
	}
	defer tx.Rollback(ctx)
	event := events.NewEvent("payment.payment.failed", "payment", payment.ID, map[string]interface{}{
		"payment_id": payment.ID, "order_id": payment.OrderID, "reason": reason,
	})
	if err := outbox.WriteToOutbox(ctx, tx, event); err != nil {
		slog.Error("publish payment failed: outbox", "payment_id", payment.ID, "error", err)
		return
	}
	tx.Commit(ctx)
}

// Commission split: the shop receives 95% of each payment, the platform keeps 5%.
func shopShare(amount int64) int64     { return amount * 95 / 100 }
func platformShare(amount int64) int64 { return amount - shopShare(amount) }

// paymentEntries are the double-entry postings for a completed payment: the customer
// is debited the full amount, which is credited to the shop and the platform.
func paymentEntries(p *domain.Payment, customerID string) []*domain.LedgerEntry {
	return []*domain.LedgerEntry{
		domain.NewLedgerEntry(p.ID, customerAccount(customerID), domain.Debit, p.Amount, "order payment"),
		domain.NewLedgerEntry(p.ID, shopAccount(p.ShopID), domain.Credit, shopShare(p.Amount), "order revenue (95%)"),
		domain.NewLedgerEntry(p.ID, "platform:revenue", domain.Credit, platformShare(p.Amount), "platform commission (5%)"),
	}
}

// refundEntries reverse paymentEntries.
func refundEntries(p *domain.Payment, customerID string) []*domain.LedgerEntry {
	return []*domain.LedgerEntry{
		domain.NewLedgerEntry(p.ID, customerAccount(customerID), domain.Credit, p.Amount, "order refund"),
		domain.NewLedgerEntry(p.ID, shopAccount(p.ShopID), domain.Debit, shopShare(p.Amount), "order refund (95%)"),
		domain.NewLedgerEntry(p.ID, "platform:revenue", domain.Debit, platformShare(p.Amount), "commission refund (5%)"),
	}
}

func customerAccount(customerID string) string { return "customer:" + customerID }

func shopAccount(shopID string) string { return "shop:" + shopID }
