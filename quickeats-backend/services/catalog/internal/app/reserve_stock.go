package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/quickeats/services/catalog/internal/domain"
	"github.com/quickeats/services/catalog/internal/port"
)

const (
	// reservationTTL is how long an unpaid order holds its stock; the order service
	// cancels PENDING_PAYMENT orders after the same time.
	reservationTTL = 15 * time.Minute
	// expiryGrace gives the order service time to cancel and release first, so the
	// catalog sweeper only cleans up what it missed.
	expiryGrace = 15 * time.Minute

	reservationHeld      = "HELD"
	reservationCommitted = "COMMITTED"
)

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
	// Idempotent per order: a redelivered order.placed event makes the saga ask again,
	// and reserving twice would take the stock twice.
	if existing, err := uc.existingReservation(ctx, orderID); err != nil {
		return nil, nil, err
	} else if existing != nil {
		return existing, nil, nil
	}

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

// Release returns an order's reserved stock. With no reservation IDs it releases every
// reservation the order holds, which is what the order service uses.
//   - HELD rows only took stock from Redis, so Redis gets it back.
//   - COMMITTED rows (confirmed order, later cancelled or rejected) also took it from
//     menu_items.stock, so both are restored.
//
// Deleting the rows makes it idempotent: a repeated release finds nothing to return.
func (uc *ReserveStockUseCase) Release(ctx context.Context, orderID string, reservationIDs []string) (int, error) {
	if len(reservationIDs) == 0 {
		return uc.deleteAndRestore(ctx,
			`DELETE FROM stock_reservations WHERE order_id = $1 RETURNING menu_item_id, quantity, status`, orderID)
	}
	return uc.deleteAndRestore(ctx,
		`DELETE FROM stock_reservations WHERE order_id = $1 AND id = ANY($2::uuid[]) RETURNING menu_item_id, quantity, status`,
		orderID, reservationIDs)
}

// Commit makes an order's HELD reservations final once the order is confirmed: the
// sold quantity is taken from menu_items.stock, which is what the shop and customers
// see. Redis already reflects it from the reservation. Only HELD rows are touched,
// so a repeated commit changes nothing. Returns how many reservations were committed.
func (uc *ReserveStockUseCase) Commit(ctx context.Context, orderID string) (int, error) {
	tx, err := uc.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx,
		`UPDATE stock_reservations SET status = $1 WHERE order_id = $2 AND status = $3 RETURNING menu_item_id, quantity`,
		reservationCommitted, orderID, reservationHeld)
	if err != nil {
		return 0, fmt.Errorf("commit reservations for %s: %w", orderID, err)
	}
	items, err := scanItems(rows)
	if err != nil {
		return 0, fmt.Errorf("commit reservations for %s: %w", orderID, err)
	}

	for _, item := range items {
		if _, err := tx.Exec(ctx,
			`UPDATE menu_items SET stock = GREATEST(stock - $1, 0), available = stock - $1 > 0, updated_at = now() WHERE id = $2`,
			item.Quantity, item.MenuItemID); err != nil {
			return 0, fmt.Errorf("take stock for %s: %w", item.MenuItemID, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return len(items), nil
}

// ReleaseExpired returns stock held by reservations that were never confirmed. The
// order service cancels unpaid orders after reservationTTL and releases their stock
// itself; this is the backstop for reservations it never cleaned up (e.g. it was down).
func (uc *ReserveStockUseCase) ReleaseExpired(ctx context.Context) (int, error) {
	return uc.deleteAndRestore(ctx,
		`DELETE FROM stock_reservations WHERE status = $1 AND expires_at < $2 RETURNING menu_item_id, quantity, status`,
		reservationHeld, time.Now().UTC().Add(-expiryGrace))
}

// RunExpirySweeper calls ReleaseExpired every interval until ctx is cancelled.
func (uc *ReserveStockUseCase) RunExpirySweeper(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n, err := uc.ReleaseExpired(ctx)
			if err != nil {
				slog.Error("stock expiry sweep failed", "error", err)
			} else if n > 0 {
				slog.Info("released expired stock reservations", "count", n)
			}
		}
	}
}

// deleteAndRestore deletes reservations with query (which must RETURNING menu_item_id,
// quantity, status), adds COMMITTED quantities back to menu_items.stock in the same
// transaction, then returns every quantity to the Redis counter.
func (uc *ReserveStockUseCase) deleteAndRestore(ctx context.Context, query string, args ...any) (int, error) {
	tx, err := uc.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("release reservations: %w", err)
	}
	type released struct {
		item   port.StockReservationItem
		status string
	}
	var all []released
	for rows.Next() {
		var r released
		if err := rows.Scan(&r.item.MenuItemID, &r.item.Quantity, &r.status); err != nil {
			rows.Close()
			return 0, fmt.Errorf("scan released reservation: %w", err)
		}
		all = append(all, r)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("release reservations: %w", err)
	}

	for _, r := range all {
		if r.status != reservationCommitted {
			continue
		}
		if _, err := tx.Exec(ctx,
			`UPDATE menu_items SET stock = stock + $1, available = true, updated_at = now() WHERE id = $2`,
			r.item.Quantity, r.item.MenuItemID); err != nil {
			return 0, fmt.Errorf("restore stock for %s: %w", r.item.MenuItemID, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}

	// After the commit, so a rolled-back release never inflates the counter.
	for _, r := range all {
		uc.stockCache.IncrementStock(ctx, r.item.MenuItemID, r.item.Quantity)
	}
	return len(all), nil
}

// scanItems reads (menu_item_id, quantity) rows and closes them.
func scanItems(rows pgx.Rows) ([]port.StockReservationItem, error) {
	defer rows.Close()
	var items []port.StockReservationItem
	for rows.Next() {
		var item port.StockReservationItem
		if err := rows.Scan(&item.MenuItemID, &item.Quantity); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// existingReservation returns the reservations already held by the order, or nil if none.
func (uc *ReserveStockUseCase) existingReservation(ctx context.Context, orderID string) (*ReserveResult, error) {
	rows, err := uc.db.Query(ctx,
		`SELECT id, menu_item_id, quantity, expires_at FROM stock_reservations WHERE order_id = $1`, orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("load reservations for %s: %w", orderID, err)
	}
	defer rows.Close()

	var result ReserveResult
	for rows.Next() {
		var res port.StockReservationResult
		var expiresAt time.Time
		if err := rows.Scan(&res.ReservationID, &res.MenuItemID, &res.Quantity, &expiresAt); err != nil {
			return nil, fmt.Errorf("scan reservation: %w", err)
		}
		result.Reservations = append(result.Reservations, res)
		result.ExpiresAt = expiresAt
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("load reservations for %s: %w", orderID, err)
	}
	if len(result.Reservations) == 0 {
		return nil, nil
	}
	return &result, nil
}

func (uc *ReserveStockUseCase) rollbackRedis(ctx context.Context, items []port.StockReservationItem) {
	for _, item := range items {
		uc.stockCache.IncrementStock(ctx, item.MenuItemID, item.Quantity)
	}
}
