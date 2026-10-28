package app

import (
	"context"
	"encoding/json"
	"fmt"
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

// orderEventData holds the fields the order events carry (each event sets a subset).
type orderEventData struct {
	OrderID     string `json:"order_id"`
	CustomerID  string `json:"customer_id"`
	ShopID      string `json:"shop_id"`
	PickupCode  string `json:"pickup_code"`
	TotalAmount int64  `json:"total_amount"`
	Currency    string `json:"currency"`
	Reason      string `json:"reason"`
	NewStatus   string `json:"new_status"`
}

func decodeOrderEvent(event events.Event) (orderEventData, error) {
	var data orderEventData
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return data, fmt.Errorf("re-marshal payload: %w", err)
	}
	if err := json.Unmarshal(payload, &data); err != nil {
		return data, fmt.Errorf("decode %s payload: %w", event.EventType, err)
	}
	return data, nil
}

// statusMessage is what the customer app receives for every change to their order.
func statusMessage(data orderEventData, newStatus string) domain.WSMessage {
	msg := map[string]interface{}{"orderId": data.OrderID, "newStatus": newStatus}
	if data.PickupCode != "" {
		msg["pickupCode"] = data.PickupCode
	}
	if data.Reason != "" {
		msg["reason"] = data.Reason
	}
	return domain.WSMessage{Type: string(domain.TypeOrderStatus), Data: msg}
}

// HandleOrderPlaced tells the customer their order was received. The shop is told only
// once payment succeeds (HandleOrderConfirmed), so unpaid orders never reach its queue.
func (s *NotificationService) HandleOrderPlaced(ctx context.Context, event events.Event) error {
	data, err := decodeOrderEvent(event)
	if err != nil {
		return err
	}

	s.hub.SendToCustomer(data.CustomerID, statusMessage(data, "PENDING_PAYMENT"))

	slog.Info("notification: order placed", "customer_id", data.CustomerID, "order_id", data.OrderID)
	return nil
}

// HandleOrderConfirmed tells the shop about a new paid order and the customer that it is confirmed.
func (s *NotificationService) HandleOrderConfirmed(ctx context.Context, event events.Event) error {
	data, err := decodeOrderEvent(event)
	if err != nil {
		return err
	}

	s.hub.SendToShop(data.ShopID, domain.WSMessage{
		Type: string(domain.TypeNewOrder),
		Data: map[string]interface{}{
			"orderId":     data.OrderID,
			"pickupCode":  data.PickupCode,
			"totalAmount": data.TotalAmount,
			"currency":    data.Currency,
		},
	})
	s.hub.SendToCustomer(data.CustomerID, statusMessage(data, "CONFIRMED"))

	slog.Info("notification: order confirmed", "shop_id", data.ShopID, "order_id", data.OrderID)
	return nil
}

// HandleOrderStatusChanged relays shop workflow updates (accepted, preparing, ready, picked up)
// to the customer, and to the shop's other devices.
func (s *NotificationService) HandleOrderStatusChanged(ctx context.Context, event events.Event) error {
	data, err := decodeOrderEvent(event)
	if err != nil {
		return err
	}

	s.hub.SendToShop(data.ShopID, domain.WSMessage{
		Type: string(domain.TypeOrderStatus),
		Data: map[string]interface{}{"orderId": data.OrderID, "newStatus": data.NewStatus},
	})
	s.hub.SendToCustomer(data.CustomerID, statusMessage(data, data.NewStatus))

	slog.Info("notification: order status changed", "order_id", data.OrderID, "new_status", data.NewStatus)
	return nil
}

// HandleOrderCancelled tells the shop to drop the order and the customer that it was cancelled.
func (s *NotificationService) HandleOrderCancelled(ctx context.Context, event events.Event) error {
	data, err := decodeOrderEvent(event)
	if err != nil {
		return err
	}

	s.hub.SendToShop(data.ShopID, domain.WSMessage{
		Type: string(domain.TypeOrderCancelled),
		Data: map[string]interface{}{"orderId": data.OrderID, "reason": data.Reason},
	})
	s.hub.SendToCustomer(data.CustomerID, statusMessage(data, "CANCELLED"))

	slog.Info("notification: order cancelled", "order_id", data.OrderID, "reason", data.Reason)
	return nil
}

// HandleOrderRejected tells the customer the shop rejected the order, and syncs the shop's other devices.
func (s *NotificationService) HandleOrderRejected(ctx context.Context, event events.Event) error {
	data, err := decodeOrderEvent(event)
	if err != nil {
		return err
	}

	s.hub.SendToShop(data.ShopID, domain.WSMessage{
		Type: string(domain.TypeOrderStatus),
		Data: map[string]interface{}{"orderId": data.OrderID, "newStatus": "REJECTED"},
	})
	s.hub.SendToCustomer(data.CustomerID, statusMessage(data, "REJECTED"))

	slog.Info("notification: order rejected", "order_id", data.OrderID, "reason", data.Reason)
	return nil
}
