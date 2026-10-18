package app

import (
	"context"
	"encoding/json"
	"log/slog"

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

// HandleOrderPlaced starts the saga: reserve stock in catalog, then request payment.
func (s *SagaOrchestrator) HandleOrderPlaced(ctx context.Context, event events.Event) error {
	payload, _ := json.Marshal(event.Payload)
	var data struct {
		OrderID string `json:"order_id"`
	}
	json.Unmarshal(payload, &data)

	order, err := s.orderRepo.GetByID(ctx, data.OrderID)
	if err != nil || order == nil {
		slog.Error("saga: order not found for order placed", "order_id", data.OrderID)
		return nil
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
	payload, _ := json.Marshal(event.Payload)
	var data struct {
		OrderID      string `json:"order_id"`
		Reservations []struct {
			MenuItemID    string `json:"menu_item_id"`
			ReservationID string `json:"reservation_id"`
		} `json:"reservations"`
	}
	json.Unmarshal(payload, &data)

	order, err := s.orderRepo.GetByID(ctx, data.OrderID)
	if err != nil || order == nil {
		slog.Error("saga: order not found for stock reserved", "order_id", data.OrderID)
		return nil
	}

	s.requestPayment(ctx, order)
	return nil
}

func (s *SagaOrchestrator) requestPayment(ctx context.Context, order *domain.Order) {
	_, err := s.paymentClient.CreatePaymentIntent(ctx, port.CreatePaymentIntentRequest{
		OrderID:    order.ID,
		CustomerID: order.CustomerID,
		Amount:     order.TotalAmount,
		Currency:   order.Currency,
		Method:     order.PaymentMethod,
	})
	if err != nil {
		slog.Error("saga: payment intent failed, compensating", "order_id", order.ID, "error", err)
		s.compensateCancel(ctx, order, "payment_failed")
		return
	}

	slog.Info("saga: payment intent created", "order_id", order.ID)
}

// HandlePaymentCompleted is called when payment.payment.completed event arrives.
func (s *SagaOrchestrator) HandlePaymentCompleted(ctx context.Context, event events.Event) error {
	payload, _ := json.Marshal(event.Payload)
	var data struct {
		OrderID string `json:"order_id"`
	}
	json.Unmarshal(payload, &data)

	order, err := s.orderRepo.GetByID(ctx, data.OrderID)
	if err != nil || order == nil {
		return nil
	}

	if err := order.TransitionTo(domain.StatusConfirmed); err != nil {
		slog.Error("saga: transition to confirmed failed", "order_id", order.ID, "error", err)
		return nil
	}

	// Write confirmed status + event in single tx
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tx.Exec(ctx, `UPDATE orders SET status=$1, pickup_code=$2, updated_at=now() WHERE id=$3`,
		order.Status, order.PickupCode, order.ID)

	confirmedEvent := events.NewEvent("order.order.confirmed", "order", order.ID, domain.OrderConfirmedPayload{
		OrderID: order.ID, CustomerID: order.CustomerID, ShopID: order.ShopID,
		PickupCode: order.PickupCode, TotalAmount: order.TotalAmount, Currency: order.Currency,
	})
	outbox.WriteToOutbox(ctx, tx, confirmedEvent)
	tx.Commit(ctx)

	slog.Info("saga: order confirmed", "order_id", order.ID, "pickup_code", order.PickupCode)
	return nil
}

// HandlePaymentFailed triggers compensation: release stock + cancel order.
func (s *SagaOrchestrator) HandlePaymentFailed(ctx context.Context, event events.Event) error {
	payload, _ := json.Marshal(event.Payload)
	var data struct {
		OrderID string `json:"order_id"`
	}
	json.Unmarshal(payload, &data)

	order, _ := s.orderRepo.GetByID(ctx, data.OrderID)
	if order == nil {
		return nil
	}

	s.compensateCancel(ctx, order, "payment_failed")
	return nil
}

// HandleStockInsufficient triggers compensation: cancel order.
func (s *SagaOrchestrator) HandleStockInsufficient(ctx context.Context, event events.Event) error {
	payload, _ := json.Marshal(event.Payload)
	var data struct {
		OrderID string `json:"order_id"`
	}
	json.Unmarshal(payload, &data)

	order, _ := s.orderRepo.GetByID(ctx, data.OrderID)
	if order == nil {
		return nil
	}

	s.compensateCancel(ctx, order, "stock_insufficient")
	return nil
}

func (s *SagaOrchestrator) compensateCancel(ctx context.Context, order *domain.Order, reason string) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		slog.Error("saga compensate: begin tx failed", "error", err)
		return
	}
	defer tx.Rollback(ctx)

	tx.Exec(ctx, `UPDATE orders SET status=$1, updated_at=now() WHERE id=$2`, domain.StatusCancelled, order.ID)

	cancelEvent := events.NewEvent("order.order.cancelled", "order", order.ID, domain.OrderCancelledPayload{
		OrderID: order.ID, CustomerID: order.CustomerID, ShopID: order.ShopID,
		Reason: reason, TotalAmount: order.TotalAmount, Currency: order.Currency,
	})
	outbox.WriteToOutbox(ctx, tx, cancelEvent)
	tx.Commit(ctx)

	slog.Info("saga: order cancelled (compensation)", "order_id", order.ID, "reason", reason)
}
