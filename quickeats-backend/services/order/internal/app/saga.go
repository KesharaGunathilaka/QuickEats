package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/quickeats/pkg/events"
	"github.com/quickeats/pkg/outbox"
	"github.com/quickeats/services/order/internal/domain"
	"github.com/quickeats/services/order/internal/port"
)

// SagaOrchestrator handles event-driven saga coordination for the order lifecycle.
type SagaOrchestrator struct {
	db            *pgxpool.Pool
	orderRepo     port.OrderRepository
	catalogClient port.CatalogClient
	paymentClient port.PaymentClient
}

func NewSagaOrchestrator(db *pgxpool.Pool, orderRepo port.OrderRepository, catalogClient port.CatalogClient, paymentClient port.PaymentClient) *SagaOrchestrator {
	return &SagaOrchestrator{db: db, orderRepo: orderRepo, catalogClient: catalogClient, paymentClient: paymentClient}
}

// loadOrder decodes the order_id from a saga event and loads that order.
//   - Malformed payload or missing order_id: logged and skipped (nil order, nil error).
//     Retrying cannot fix a bad payload, so returning an error would only stall the topic.
//   - Database error: returned, so the consumer retries the event.
//   - Unknown order: logged and skipped.
func (s *SagaOrchestrator) loadOrder(ctx context.Context, event events.Event) (*domain.Order, error) {
	var data struct {
		OrderID string `json:"order_id"`
	}
	payload, err := json.Marshal(event.Payload)
	if err == nil {
		err = json.Unmarshal(payload, &data)
	}
	if err == nil && data.OrderID == "" {
		err = errors.New("missing order_id")
	}
	if err != nil {
		slog.Error("saga: malformed event payload, skipping",
			"event_type", event.EventType, "event_id", event.EventID, "error", err)
		return nil, nil
	}

	order, err := s.orderRepo.GetByID(ctx, data.OrderID)
	if err != nil {
		return nil, fmt.Errorf("saga: load order %s for %s: %w", data.OrderID, event.EventType, err)
	}
	if order == nil {
		slog.Error("saga: order not found, skipping", "event_type", event.EventType, "order_id", data.OrderID)
	}
	return order, nil
}

// HandleOrderPlaced starts the saga: reserve stock in catalog, then request payment.
func (s *SagaOrchestrator) HandleOrderPlaced(ctx context.Context, event events.Event) error {
	order, err := s.loadOrder(ctx, event)
	if err != nil || order == nil {
		return err
	}
	if order.Status != domain.StatusPendingPayment {
		return nil // already progressed (redelivered event)
	}

	var items []port.ReserveStockItem
	for _, item := range order.Items {
		items = append(items, port.ReserveStockItem{MenuItemID: item.MenuItemID, Quantity: item.Quantity})
	}
	if _, err := s.catalogClient.ReserveStock(ctx, port.ReserveStockRequest{OrderID: order.ID, Items: items}); err != nil {
		slog.Error("saga: stock reservation failed, compensating", "order_id", order.ID, "error", err)
		s.compensateCancel(ctx, order, "stock_insufficient")
		return nil
	}

	s.requestPayment(ctx, order)
	return nil
}

// HandleStockReserved is called when catalog.stock.reserved event arrives.
func (s *SagaOrchestrator) HandleStockReserved(ctx context.Context, event events.Event) error {
	order, err := s.loadOrder(ctx, event)
	if err != nil || order == nil {
		return err
	}

	s.requestPayment(ctx, order)
	return nil
}

func (s *SagaOrchestrator) requestPayment(ctx context.Context, order *domain.Order) {
	_, err := s.paymentClient.CreatePaymentIntent(ctx, port.CreatePaymentIntentRequest{
		OrderID:    order.ID,
		CustomerID: order.CustomerID,
		ShopID:     order.ShopID,
		Amount:     order.TotalAmount,
		Currency:   order.Currency,
		Method:     order.PaymentMethod,
	})
	if err != nil {
		// A declined card carries its reason (no_card, insufficient_funds), which is
		// stored on the cancelled order so the app can tell the customer what to fix.
		reason := port.DeclineOther
		var declined *port.PaymentDeclinedError
		if errors.As(err, &declined) {
			reason = declined.Reason
		}
		slog.Error("saga: payment failed, compensating", "order_id", order.ID, "reason", reason, "error", err)
		s.compensateCancel(ctx, order, reason)
		return
	}

	slog.Info("saga: payment intent created", "order_id", order.ID)
}

// HandlePaymentCompleted is called when payment.payment.completed event arrives.
func (s *SagaOrchestrator) HandlePaymentCompleted(ctx context.Context, event events.Event) error {
	order, err := s.loadOrder(ctx, event)
	if err != nil || order == nil {
		return err
	}

	prevStatus := order.Status
	if err := order.TransitionTo(domain.StatusConfirmed); err != nil {
		slog.Error("saga: transition to confirmed failed", "order_id", order.ID, "error", err)
		// Paid for an order that was already cancelled (e.g. payment timeout): give the money back.
		s.refundIfClosed(ctx, order)
		return nil
	}

	// Take the sold items from the shop's stock before confirming. Commit is
	// idempotent, so if the confirm below fails the retried event commits nothing new.
	committed, err := s.catalogClient.CommitStock(ctx, order.ID)
	if err != nil {
		return fmt.Errorf("saga confirm %s: commit stock: %w", order.ID, err)
	}
	if committed == 0 {
		slog.Warn("saga confirm: no held reservations to commit", "order_id", order.ID)
	}

	// Write confirmed status + event in single tx. Errors are returned so the
	// consumer retries; the status check above makes a retry after commit a no-op.
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("saga confirm %s: begin tx: %w", order.ID, err)
	}
	defer tx.Rollback(ctx)

	// Conditional on the status we read: if the order was cancelled meanwhile (e.g.
	// payment timeout), it stays cancelled, and that cancel releases the stock we
	// just committed.
	tag, err := tx.Exec(ctx, `UPDATE orders SET status=$1, pickup_code=$2, updated_at=now() WHERE id=$3 AND status=$4`,
		order.Status, order.PickupCode, order.ID, prevStatus)
	if err != nil {
		return fmt.Errorf("saga confirm %s: update order: %w", order.ID, err)
	}
	if tag.RowsAffected() == 0 {
		slog.Warn("saga confirm: order changed status meanwhile, not confirming", "order_id", order.ID)
		if current, err := s.orderRepo.GetByID(ctx, order.ID); err == nil && current != nil {
			s.refundIfClosed(ctx, current)
		}
		return nil
	}

	confirmedEvent := events.NewEvent("order.order.confirmed", "order", order.ID, domain.OrderConfirmedPayload{
		OrderID: order.ID, CustomerID: order.CustomerID, ShopID: order.ShopID,
		PickupCode: order.PickupCode, TotalAmount: order.TotalAmount, Currency: order.Currency,
	})
	if err := outbox.WriteToOutbox(ctx, tx, confirmedEvent); err != nil {
		return fmt.Errorf("saga confirm %s: write outbox: %w", order.ID, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("saga confirm %s: commit: %w", order.ID, err)
	}

	slog.Info("saga: order confirmed", "order_id", order.ID, "pickup_code", order.PickupCode)
	return nil
}

// HandlePaymentFailed triggers compensation: release stock + cancel order.
func (s *SagaOrchestrator) HandlePaymentFailed(ctx context.Context, event events.Event) error {
	order, err := s.loadOrder(ctx, event)
	if err != nil || order == nil {
		return err
	}

	s.compensateCancel(ctx, order, declineReason(event))
	return nil
}

// declineReason reads the reason code from a payment.failed event.
func declineReason(event events.Event) string {
	var data struct {
		Reason string `json:"reason"`
	}
	payload, err := json.Marshal(event.Payload)
	if err == nil {
		err = json.Unmarshal(payload, &data)
	}
	switch data.Reason {
	case port.DeclineNoCard, port.DeclineInsufficientFunds:
		return data.Reason
	default:
		return port.DeclineOther
	}
}

// refundIfClosed refunds the payment of an order that ended up cancelled or rejected.
// Refunding is idempotent in the payment service, so repeating it is harmless.
func (s *SagaOrchestrator) refundIfClosed(ctx context.Context, order *domain.Order) {
	if order.Status != domain.StatusCancelled && order.Status != domain.StatusRejected {
		return
	}
	if err := s.paymentClient.RefundOrder(ctx, order.ID, order.CustomerID); err != nil {
		slog.Error("saga: refund failed", "order_id", order.ID, "error", err)
		return
	}
	slog.Info("saga: refunded payment for closed order", "order_id", order.ID)
}

// HandleStockInsufficient triggers compensation: cancel order.
func (s *SagaOrchestrator) HandleStockInsufficient(ctx context.Context, event events.Event) error {
	order, err := s.loadOrder(ctx, event)
	if err != nil || order == nil {
		return err
	}

	s.compensateCancel(ctx, order, "stock_insufficient")
	return nil
}

// compensateCancel undoes the saga's completed steps: it cancels the order and
// releases the stock reserved for it in catalog.
func (s *SagaOrchestrator) compensateCancel(ctx context.Context, order *domain.Order, reason string) {
	if order.Status == domain.StatusCancelled {
		// Redelivered event: the order is already cancelled, but retry the release in
		// case it failed last time (releasing is idempotent).
		s.releaseStock(ctx, order.ID)
		return
	}
	prevStatus := order.Status
	if err := order.TransitionTo(domain.StatusCancelled); err != nil {
		slog.Error("saga compensate: order can no longer be cancelled", "order_id", order.ID, "status", order.Status, "error", err)
		return
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		slog.Error("saga compensate: begin tx failed", "error", err)
		return
	}
	defer tx.Rollback(ctx)

	// Conditional on the status we read, so a concurrent confirm and cancel can't both win.
	tag, err := tx.Exec(ctx, `UPDATE orders SET status=$1, cancel_reason=$2, updated_at=now() WHERE id=$3 AND status=$4`,
		domain.StatusCancelled, reason, order.ID, prevStatus)
	if err != nil {
		slog.Error("saga compensate: cancel order failed", "order_id", order.ID, "error", err)
		return
	}
	if tag.RowsAffected() == 0 {
		slog.Warn("saga compensate: order changed status meanwhile, not cancelling", "order_id", order.ID)
		return
	}

	cancelEvent := events.NewEvent("order.order.cancelled", "order", order.ID, domain.OrderCancelledPayload{
		OrderID: order.ID, CustomerID: order.CustomerID, ShopID: order.ShopID,
		Reason: reason, TotalAmount: order.TotalAmount, Currency: order.Currency,
	})
	if err := outbox.WriteToOutbox(ctx, tx, cancelEvent); err != nil {
		slog.Error("saga compensate: write outbox failed", "order_id", order.ID, "error", err)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		slog.Error("saga compensate: commit failed", "order_id", order.ID, "error", err)
		return
	}

	s.releaseStock(ctx, order.ID)
	slog.Info("saga: order cancelled (compensation)", "order_id", order.ID, "reason", reason)
}

// paymentTimeout matches catalog's reservation TTL: an order still unpaid after this
// is cancelled and its stock released.
const paymentTimeout = 15 * time.Minute

// CancelUnpaidOrders cancels orders that have waited longer than paymentTimeout for
// payment, releasing their stock.
func (s *SagaOrchestrator) CancelUnpaidOrders(ctx context.Context) {
	orders, err := s.orderRepo.FindTimedOutOrders(ctx, domain.StatusPendingPayment, int(paymentTimeout/time.Minute))
	if err != nil {
		slog.Error("saga: find unpaid orders failed", "error", err)
		return
	}
	for _, order := range orders {
		s.compensateCancel(ctx, order, "payment_timeout")
	}
}

// RunPaymentTimeouts calls CancelUnpaidOrders every interval until ctx is cancelled.
func (s *SagaOrchestrator) RunPaymentTimeouts(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.CancelUnpaidOrders(ctx)
		}
	}
}

// releaseStock returns every reservation held by the order. When stock was never
// reserved (e.g. stock_insufficient), catalog finds nothing to release.
func (s *SagaOrchestrator) releaseStock(ctx context.Context, orderID string) {
	if err := s.catalogClient.ReleaseStock(ctx, orderID, nil); err != nil {
		slog.Error("saga compensate: stock release failed", "order_id", orderID, "error", err)
		return
	}
	slog.Info("saga: stock released", "order_id", orderID)
}
