package domain

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID                    string      `json:"id"`
	IdempotencyKey        string      `json:"idempotencyKey"`
	CustomerID            string      `json:"customerId"`
	ShopID                string      `json:"shopId"`
	Status                OrderStatus `json:"status"`
	PickupCode            string      `json:"pickupCode,omitempty"`
	CancellationDeadline  *time.Time  `json:"cancellationDeadline,omitempty"`
	TotalAmount           int64       `json:"totalAmount"`
	Currency              string      `json:"currency"`
	Items                 []OrderItem `json:"items"`
	PaymentMethod         string      `json:"paymentMethod"`
	RejectReason          string      `json:"rejectReason,omitempty"`
	CancelReason          string      `json:"cancelReason,omitempty"` // e.g. insufficient_funds, no_card, payment_timeout
	CreatedAt             time.Time   `json:"createdAt"`
	UpdatedAt             time.Time   `json:"updatedAt"`
}

type OrderItem struct {
	ID         string `json:"id"`
	OrderID    string `json:"orderId"`
	MenuItemID string `json:"menuItemId"`
	ItemName   string `json:"itemName"`
	Quantity   int    `json:"quantity"`
	UnitPrice  int64  `json:"unitPrice"`
	Subtotal   int64  `json:"subtotal"`
}

func NewOrder(customerID, shopID, idempotencyKey, paymentMethod string, items []OrderItem) *Order {
	now := time.Now().UTC()
	var total int64
	for i := range items {
		items[i].ID = uuid.New().String()
		items[i].Subtotal = items[i].UnitPrice * int64(items[i].Quantity)
		total += items[i].Subtotal
	}
	deadline := now.Add(5 * time.Minute)
	orderID := uuid.New().String()
	for i := range items {
		items[i].OrderID = orderID
	}
	return &Order{
		ID:                   orderID,
		IdempotencyKey:       idempotencyKey,
		CustomerID:           customerID,
		ShopID:               shopID,
		Status:               StatusPendingPayment,
		CancellationDeadline: &deadline,
		TotalAmount:          total,
		Currency:             "LKR",
		Items:                items,
		PaymentMethod:        paymentMethod,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

func GeneratePickupCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 4)
	rand.Read(b)
	var sb strings.Builder
	for _, v := range b {
		sb.WriteByte(charset[int(v)%len(charset)])
	}
	return sb.String()
}

func (o *Order) TransitionTo(status OrderStatus) error {
	if err := o.Status.Validate(status); err != nil {
		return fmt.Errorf("invalid transition: %w", err)
	}
	o.Status = status
	o.UpdatedAt = time.Now().UTC()
	if status == StatusConfirmed {
		o.PickupCode = GeneratePickupCode()
	}
	return nil
}
