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

var validTransitions = map[OrderStatus][]OrderStatus{
	StatusPendingPayment: {StatusConfirmed, StatusCancelled},
	StatusConfirmed:      {StatusAccepted, StatusCancelled, StatusRejected},
	StatusAccepted:       {StatusPreparing, StatusRejected},
	StatusPreparing:      {StatusReady},
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
