package port

import "context"

// StockReservationItem represents a single item to reserve.
type StockReservationItem struct {
	MenuItemID string
	Quantity   int
}

// StockReservationResult contains the result of a reservation.
type StockReservationResult struct {
	MenuItemID    string `json:"menuItemId"`
	ReservationID string `json:"reservationId"`
	Quantity      int    `json:"quantity"`
}

// FailedItem represents an item that could not be reserved.
type FailedItem struct {
	MenuItemID string `json:"menuItemId"`
	Requested  int    `json:"requested"`
	Available  int    `json:"available"`
}

type StockCache interface {
	// DecrementStock atomically decrements stock in Redis. Returns new count.
	DecrementStock(ctx context.Context, menuItemID string, quantity int) (int64, error)
	// IncrementStock restores stock in Redis.
	IncrementStock(ctx context.Context, menuItemID string, quantity int) error
	// GetStock returns current stock count from Redis.
	GetStock(ctx context.Context, menuItemID string) (int64, error)
	// SetStock sets the stock count in Redis (for cache warming).
	SetStock(ctx context.Context, menuItemID string, stock int) error
}
