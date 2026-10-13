package domain

type NotificationType string

const (
	TypeNewOrder      NotificationType = "NEW_ORDER"
	TypeOrderStatus   NotificationType = "ORDER_STATUS_CHANGED"
	TypeOrderReady    NotificationType = "ORDER_READY"
	TypeOrderCancelled NotificationType = "ORDER_CANCELLED"
	TypeRefundComplete NotificationType = "REFUND_COMPLETE"
)

type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
