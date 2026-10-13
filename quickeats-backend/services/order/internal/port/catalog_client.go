package port

import "context"

type ReserveStockRequest struct {
	OrderID string              `json:"orderId"`
	Items   []ReserveStockItem  `json:"items"`
}

type ReserveStockItem struct {
	MenuItemID string `json:"menuItemId"`
	Quantity   int    `json:"quantity"`
}

type ReserveStockResponse struct {
	Reservations []ReservationResult `json:"reservations"`
	ExpiresAt    string              `json:"expiresAt"`
}

type ReservationResult struct {
	MenuItemID    string `json:"menuItemId"`
	ReservationID string `json:"reservationId"`
	Quantity      int    `json:"quantity"`
}

type CatalogClient interface {
	ReserveStock(ctx context.Context, req ReserveStockRequest) (*ReserveStockResponse, error)
	ReleaseStock(ctx context.Context, orderID string, reservationIDs []string) error
}
