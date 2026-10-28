package domain

import (
	"time"
	"github.com/google/uuid"
)

type PaymentStatus string
const (
	PaymentPending   PaymentStatus = "PENDING"
	PaymentCompleted PaymentStatus = "COMPLETED"
	PaymentFailed    PaymentStatus = "FAILED"
	PaymentRefunded  PaymentStatus = "REFUNDED"
)

type Payment struct {
	ID         string        `json:"id"`
	OrderID    string        `json:"orderId"`
	ShopID     string        `json:"shopId"` // the shop being paid
	Method     string        `json:"method"`
	Status     PaymentStatus `json:"status"`
	Amount     int64         `json:"amount"`
	Currency   string        `json:"currency"`
	GatewayRef string        `json:"gatewayRef,omitempty"`
	CustomerID string        `json:"-"` // set per request for strategies; not persisted
	CreatedAt  time.Time     `json:"createdAt"`
	UpdatedAt  time.Time     `json:"updatedAt"`
}

func NewPayment(orderID, method string, amount int64, currency string) *Payment {
	now := time.Now().UTC()
	return &Payment{ID: uuid.New().String(), OrderID: orderID, Method: method, Status: PaymentPending, Amount: amount, Currency: currency, CreatedAt: now, UpdatedAt: now}
}
