package app

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/quickeats/pkg/events"
	"github.com/quickeats/services/notification/internal/domain"
)

type NotificationService struct {
	hub *Hub
}

func NewNotificationService(hub *Hub) *NotificationService {
	return &NotificationService{hub: hub}
}

// HandleOrderConfirmed sends a new order notification to the shop's WebSocket clients.
func (s *NotificationService) HandleOrderConfirmed(ctx context.Context, event events.Event) error {
	payload, _ := json.Marshal(event.Payload)
	var data struct {
		OrderID     string `json:"order_id"`
		CustomerID  string `json:"customer_id"`
		ShopID      string `json:"shop_id"`
		PickupCode  string `json:"pickup_code"`
		TotalAmount int64  `json:"total_amount"`
		Currency    string `json:"currency"`
	}
	json.Unmarshal(payload, &data)

	s.hub.SendToShop(data.ShopID, domain.WSMessage{
		Type: string(domain.TypeNewOrder),
		Data: map[string]interface{}{
			"orderId":     data.OrderID,
			"pickupCode":  data.PickupCode,
			"totalAmount": data.TotalAmount,
			"currency":    data.Currency,
		},
	})

	slog.Info("notification: order confirmed sent to shop", "shop_id", data.ShopID, "order_id", data.OrderID)
	// TODO: also send FCM push to customer
	return nil
}

// HandleOrderCancelled notifies shop about cancellation.
func (s *NotificationService) HandleOrderCancelled(ctx context.Context, event events.Event) error {
	payload, _ := json.Marshal(event.Payload)
	var data struct {
		OrderID string `json:"order_id"`
		ShopID  string `json:"shop_id"`
		Reason  string `json:"reason"`
	}
	json.Unmarshal(payload, &data)

	s.hub.SendToShop(data.ShopID, domain.WSMessage{
		Type: string(domain.TypeOrderCancelled),
		Data: map[string]interface{}{"orderId": data.OrderID, "reason": data.Reason},
	})
	return nil
}

// HandleOrderStatusChanged notifies both shop and customer.
func (s *NotificationService) HandleOrderStatusChanged(ctx context.Context, event events.Event) error {
	payload, _ := json.Marshal(event.Payload)
	var data struct {
		OrderID   string `json:"order_id"`
		ShopID    string `json:"shop_id"`
		NewStatus string `json:"new_status"`
	}
	json.Unmarshal(payload, &data)

	s.hub.SendToShop(data.ShopID, domain.WSMessage{
		Type: string(domain.TypeOrderStatus),
		Data: map[string]interface{}{"orderId": data.OrderID, "newStatus": data.NewStatus},
	})
	return nil
}
