package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/quickeats/pkg/events"
	"github.com/quickeats/pkg/outbox"
	"github.com/quickeats/services/order/internal/domain"
	"github.com/quickeats/services/order/internal/port"
)

// OrderStatusUseCase applies status changes made by the shop or the customer and
// publishes them as events, so the notification service can tell the other side.
type OrderStatusUseCase struct {
	db            *pgxpool.Pool
	catalogClient port.CatalogClient
}

func NewOrderStatusUseCase(db *pgxpool.Pool, catalogClient port.CatalogClient) *OrderStatusUseCase {
	return &OrderStatusUseCase{db: db, catalogClient: catalogClient}
}

// UpdateStatus moves an order along the shop workflow (ACCEPTED, PREPARING, READY, PICKED_UP).
// The order must already have been transitioned in memory.
func (uc *OrderStatusUseCase) UpdateStatus(ctx context.Context, order *domain.Order, oldStatus domain.OrderStatus) error {
	return uc.saveWithEvent(ctx, order, `UPDATE orders SET status=$1, updated_at=now() WHERE id=$2`,
		[]any{order.Status, order.ID},
		events.NewEvent("order.order.status_changed", "order", order.ID, domain.OrderStatusChangedPayload{
			OrderID: order.ID, ShopID: order.ShopID, CustomerID: order.CustomerID,
			OldStatus: string(oldStatus), NewStatus: string(order.Status),
		}))
}

// Reject records the shop's rejection and its reason, and returns the order's stock.
func (uc *OrderStatusUseCase) Reject(ctx context.Context, order *domain.Order, reason string) error {
	err := uc.saveWithEvent(ctx, order, `UPDATE orders SET status=$1, reject_reason=$2, updated_at=now() WHERE id=$3`,
		[]any{order.Status, reason, order.ID},
		events.NewEvent("order.order.rejected", "order", order.ID, domain.OrderCancelledPayload{
			OrderID: order.ID, CustomerID: order.CustomerID, ShopID: order.ShopID,
			Reason: reason, TotalAmount: order.TotalAmount, Currency: order.Currency,
		}))
	if err != nil {
		return err
	}
	uc.releaseStock(ctx, order.ID)
	return nil
}

// Cancel records a cancellation requested by the customer, and returns the order's stock.
func (uc *OrderStatusUseCase) Cancel(ctx context.Context, order *domain.Order, reason string) error {
	err := uc.saveWithEvent(ctx, order, `UPDATE orders SET status=$1, updated_at=now() WHERE id=$2`,
		[]any{order.Status, order.ID},
		events.NewEvent("order.order.cancelled", "order", order.ID, domain.OrderCancelledPayload{
			OrderID: order.ID, CustomerID: order.CustomerID, ShopID: order.ShopID,
			Reason: reason, TotalAmount: order.TotalAmount, Currency: order.Currency,
		}))
	if err != nil {
		return err
	}
	uc.releaseStock(ctx, order.ID)
	return nil
}

// releaseStock returns the order's reserved items to catalog. The status change is
// already committed, so a failed release is logged rather than failing the request.
func (uc *OrderStatusUseCase) releaseStock(ctx context.Context, orderID string) {
	if err := uc.catalogClient.ReleaseStock(ctx, orderID, nil); err != nil {
		slog.Error("stock release failed", "order_id", orderID, "error", err)
	}
}

// saveWithEvent writes the status change and its event in one transaction (outbox pattern).
func (uc *OrderStatusUseCase) saveWithEvent(ctx context.Context, order *domain.Order, query string, args []any, event events.Event) error {
	tx, err := uc.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("update order %s: %w", order.ID, err)
	}
	if err := outbox.WriteToOutbox(ctx, tx, event); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
