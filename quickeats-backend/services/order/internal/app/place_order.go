package app

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/quickeats/pkg/events"
	"github.com/quickeats/pkg/idempotency"
	"github.com/quickeats/pkg/outbox"
	"github.com/quickeats/services/order/internal/domain"
	"github.com/quickeats/services/order/internal/port"
)

type PlaceOrderRequest struct {
	CustomerID     string             `json:"customerId"`
	ShopID         string             `json:"shopId"`
	IdempotencyKey string             `json:"idempotencyKey"`
	PaymentMethod  string             `json:"paymentMethod"`
	Items          []domain.OrderItem `json:"items"`
}

type PlaceOrderUseCase struct {
	db          *pgxpool.Pool
	orderRepo   port.OrderRepository
	idempotency *idempotency.Checker
}

func NewPlaceOrderUseCase(db *pgxpool.Pool, orderRepo port.OrderRepository, idempotency *idempotency.Checker) *PlaceOrderUseCase {
	return &PlaceOrderUseCase{db: db, orderRepo: orderRepo, idempotency: idempotency}
}

func (uc *PlaceOrderUseCase) Execute(ctx context.Context, req PlaceOrderRequest) (*domain.Order, error) {
	// 1. Idempotency check
	isNew, err := uc.idempotency.Check(ctx, req.IdempotencyKey, 24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("idempotency check: %w", err)
	}
	if !isNew {
		existing, err := uc.orderRepo.GetByIdempotencyKey(ctx, req.IdempotencyKey)
		if err == nil && existing != nil {
			return existing, nil
		}
		return nil, domain.ErrDuplicateOrder
	}

	// 2. Create order
	order := domain.NewOrder(req.CustomerID, req.ShopID, req.IdempotencyKey, req.PaymentMethod, req.Items)

	// 3. Write order + outbox event in single transaction
	tx, err := uc.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert order
	_, err = tx.Exec(ctx,
		`INSERT INTO orders (id, idempotency_key, customer_id, shop_id, status, pickup_code, cancellation_deadline, total_amount, currency, payment_method, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		order.ID, order.IdempotencyKey, order.CustomerID, order.ShopID, order.Status, order.PickupCode,
		order.CancellationDeadline, order.TotalAmount, order.Currency, order.PaymentMethod, order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert order: %w", err)
	}

	// Insert order items
	for _, item := range order.Items {
		_, err = tx.Exec(ctx,
			`INSERT INTO order_items (id, order_id, menu_item_id, item_name, quantity, unit_price, subtotal)
			 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			item.ID, item.OrderID, item.MenuItemID, item.ItemName, item.Quantity, item.UnitPrice, item.Subtotal,
		)
		if err != nil {
			return nil, fmt.Errorf("insert order item: %w", err)
		}
	}

	// Build event payload
	var itemPayloads []domain.OrderItemPayload
	for _, item := range order.Items {
		itemPayloads = append(itemPayloads, domain.OrderItemPayload{
			MenuItemID: item.MenuItemID, ItemName: item.ItemName,
			Quantity: item.Quantity, UnitPrice: item.UnitPrice, Subtotal: item.Subtotal,
		})
	}

	event := events.NewEvent("order.order.placed", "order", order.ID, domain.OrderPlacedPayload{
		OrderID: order.ID, CustomerID: order.CustomerID, ShopID: order.ShopID,
		Items: itemPayloads, TotalAmount: order.TotalAmount, Currency: order.Currency,
		IdempotencyKey: order.IdempotencyKey, PaymentMethod: order.PaymentMethod,
	})

	if err := outbox.WriteToOutbox(ctx, tx, event); err != nil {
		return nil, fmt.Errorf("write outbox: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return order, nil
}
