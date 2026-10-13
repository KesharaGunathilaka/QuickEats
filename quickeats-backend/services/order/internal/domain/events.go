package domain

// Event payloads (snake_case for events as per convention)

type OrderPlacedPayload struct {
	OrderID        string             `json:"order_id"`
	CustomerID     string             `json:"customer_id"`
	ShopID         string             `json:"shop_id"`
	Items          []OrderItemPayload `json:"items"`
	TotalAmount    int64              `json:"total_amount"`
	Currency       string             `json:"currency"`
	IdempotencyKey string             `json:"idempotency_key"`
	PaymentMethod  string             `json:"payment_method"`
}

type OrderItemPayload struct {
	MenuItemID string `json:"menu_item_id"`
	ItemName   string `json:"item_name"`
	Quantity   int    `json:"quantity"`
	UnitPrice  int64  `json:"unit_price"`
	Subtotal   int64  `json:"subtotal"`
}

type OrderConfirmedPayload struct {
	OrderID     string `json:"order_id"`
	CustomerID  string `json:"customer_id"`
	ShopID      string `json:"shop_id"`
	PickupCode  string `json:"pickup_code"`
	TotalAmount int64  `json:"total_amount"`
	Currency    string `json:"currency"`
}

type OrderCancelledPayload struct {
	OrderID     string `json:"order_id"`
	CustomerID  string `json:"customer_id"`
	ShopID      string `json:"shop_id"`
	Reason      string `json:"reason"`
	TotalAmount int64  `json:"total_amount"`
	Currency    string `json:"currency"`
}

type OrderStatusChangedPayload struct {
	OrderID    string `json:"order_id"`
	ShopID     string `json:"shop_id"`
	CustomerID string `json:"customer_id"`
	OldStatus  string `json:"old_status"`
	NewStatus  string `json:"new_status"`
}
