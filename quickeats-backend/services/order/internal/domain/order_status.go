package domain

import "fmt"

type OrderStatus string

const (
	StatusPendingPayment OrderStatus = "PENDING_PAYMENT"
	StatusConfirmed      OrderStatus = "CONFIRMED"
	StatusAccepted       OrderStatus = "ACCEPTED"
	StatusPreparing      OrderStatus = "PREPARING"
	StatusReady          OrderStatus = "READY"
	StatusPickedUp       OrderStatus = "PICKED_UP"
	StatusCancelled      OrderStatus = "CANCELLED"
	StatusRejected       OrderStatus = "REJECTED"
	StatusRefunded       OrderStatus = "REFUNDED"
)

// CONFIRMED means paid and waiting for the shop: the customer may still cancel and the
// shop may accept or reject. Once ACCEPTED the order can only be completed (picked up).
// PREPARING and READY are no longer used by the shop app; they stay so older orders
// in those states can still be completed.
var validTransitions = map[OrderStatus][]OrderStatus{
	StatusPendingPayment: {StatusConfirmed, StatusCancelled},
	StatusConfirmed:      {StatusAccepted, StatusCancelled, StatusRejected},
	StatusAccepted:       {StatusPickedUp},
	StatusPreparing:      {StatusReady, StatusPickedUp},
	StatusReady:          {StatusPickedUp},
	StatusCancelled:      {StatusRefunded},
	StatusRejected:       {StatusRefunded},
}

func (s OrderStatus) CanTransitionTo(target OrderStatus) bool {
	allowed, ok := validTransitions[s]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == target {
			return true
		}
	}
	return false
}

func (s OrderStatus) Validate(target OrderStatus) error {
	if !s.CanTransitionTo(target) {
		return fmt.Errorf("cannot transition from %s to %s", s, target)
	}
	return nil
}

func (s OrderStatus) IsTerminal() bool {
	return s == StatusPickedUp || s == StatusRefunded
}
