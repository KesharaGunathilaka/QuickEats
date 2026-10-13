package domain

import (
	"time"

	"github.com/google/uuid"
)

type StockReservation struct {
	ID         string    `json:"id"`
	OrderID    string    `json:"orderId"`
	MenuItemID string    `json:"menuItemId"`
	Quantity   int       `json:"quantity"`
	ExpiresAt  time.Time `json:"expiresAt"`
	CreatedAt  time.Time `json:"createdAt"`
}

func NewStockReservation(orderID, menuItemID string, quantity int, ttl time.Duration) *StockReservation {
	now := time.Now().UTC()
	return &StockReservation{
		ID:         uuid.New().String(),
		OrderID:    orderID,
		MenuItemID: menuItemID,
		Quantity:   quantity,
		ExpiresAt:  now.Add(ttl),
		CreatedAt:  now,
	}
}
