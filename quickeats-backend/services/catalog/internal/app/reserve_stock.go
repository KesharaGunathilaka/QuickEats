package app

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/quickeats/services/catalog/internal/domain"
	"github.com/quickeats/services/catalog/internal/port"
)

const reservationTTL = 15 * time.Minute

type ReserveStockUseCase struct {
	db         *pgxpool.Pool
	stockCache port.StockCache
}

func NewReserveStockUseCase(db *pgxpool.Pool, stockCache port.StockCache) *ReserveStockUseCase {
	return &ReserveStockUseCase{db: db, stockCache: stockCache}
}

type ReserveResult struct {
	Reservations []port.StockReservationResult `json:"reservations"`
	ExpiresAt    time.Time                     `json:"expiresAt"`
}

func (uc *ReserveStockUseCase) Reserve(ctx context.Context, orderID string, items []port.StockReservationItem) (*ReserveResult, []port.FailedItem, error) {
	// Phase 1: Atomic Redis decrements
	var decremented []port.StockReservationItem
	var failed []port.FailedItem

	for _, item := range items {
		newCount, err := uc.stockCache.DecrementStock(ctx, item.MenuItemID, item.Quantity)
		if err != nil {
			// Rollback all decremented so far
			uc.rollbackRedis(ctx, decremented)
			return nil, nil, fmt.Errorf("redis decrement: %w", err)
		}
		if newCount < 0 {
			// Rollback this item
			uc.stockCache.IncrementStock(ctx, item.MenuItemID, item.Quantity)
			avail := int(newCount) + item.Quantity
			if avail < 0 {
				avail = 0
			}
			failed = append(failed, port.FailedItem{
				MenuItemID: item.MenuItemID,
				Requested:  item.Quantity,
				Available:  avail,
			})
		} else {
			decremented = append(decremented, item)
		}
	}

	if len(failed) > 0 {
		uc.rollbackRedis(ctx, decremented)
		return nil, failed, domain.ErrStockInsufficient
	}

	// Phase 2: Write reservation rows to PostgreSQL
	expiresAt := time.Now().UTC().Add(reservationTTL)
	var reservations []port.StockReservationResult

	tx, err := uc.db.Begin(ctx)
	if err != nil {
		uc.rollbackRedis(ctx, decremented)
		return nil, nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, item := range decremented {
		res := domain.NewStockReservation(orderID, item.MenuItemID, item.Quantity, reservationTTL)
		_, err := tx.Exec(ctx,
			`INSERT INTO stock_reservations (id, order_id, menu_item_id, quantity, expires_at, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			res.ID, res.OrderID, res.MenuItemID, res.Quantity, res.ExpiresAt, res.CreatedAt,
		)
		if err != nil {
			uc.rollbackRedis(ctx, decremented)
			return nil, nil, fmt.Errorf("insert reservation: %w", err)
		}
		reservations = append(reservations, port.StockReservationResult{
			MenuItemID:    item.MenuItemID,
			ReservationID: res.ID,
			Quantity:      item.Quantity,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		uc.rollbackRedis(ctx, decremented)
		return nil, nil, fmt.Errorf("commit: %w", err)
	}

	return &ReserveResult{Reservations: reservations, ExpiresAt: expiresAt}, nil, nil
}

// Release returns reserved stock. With no reservation IDs it releases every reservation
// held by the order, which is what saga compensation uses. Deleting the rows makes it
// idempotent: a repeated release finds nothing left to return.
func (uc *ReserveStockUseCase) Release(ctx context.Context, orderID string, reservationIDs []string) (int, error) {
	if len(reservationIDs) == 0 {
		return uc.releaseOrder(ctx, orderID)
	}
	released := 0
	for _, resID := range reservationIDs {
		var menuItemID string
		var quantity int
		err := uc.db.QueryRow(ctx,
			`DELETE FROM stock_reservations WHERE id = $1 RETURNING menu_item_id, quantity`, resID,
		).Scan(&menuItemID, &quantity)
		if err != nil {
			continue // already released or doesn't exist
		}
		uc.stockCache.IncrementStock(ctx, menuItemID, quantity)
		released++
	}
	return released, nil
}

func (uc *ReserveStockUseCase) releaseOrder(ctx context.Context, orderID string) (int, error) {
	rows, err := uc.db.Query(ctx,
		`DELETE FROM stock_reservations WHERE order_id = $1 RETURNING menu_item_id, quantity`, orderID,
	)
	if err != nil {
		return 0, fmt.Errorf("release order %s: %w", orderID, err)
	}
	var items []port.StockReservationItem
	for rows.Next() {
		var item port.StockReservationItem
		if err := rows.Scan(&item.MenuItemID, &item.Quantity); err != nil {
			rows.Close()
			return 0, fmt.Errorf("scan released reservation: %w", err)
		}
		items = append(items, item)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("release order %s: %w", orderID, err)
	}

	for _, item := range items {
		uc.stockCache.IncrementStock(ctx, item.MenuItemID, item.Quantity)
	}
	return len(items), nil
}

func (uc *ReserveStockUseCase) rollbackRedis(ctx context.Context, items []port.StockReservationItem) {
	for _, item := range items {
		uc.stockCache.IncrementStock(ctx, item.MenuItemID, item.Quantity)
	}
}
