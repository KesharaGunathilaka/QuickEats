package domain

import "errors"

var (
	ErrOrderNotFound          = errors.New("order not found")
	ErrDuplicateOrder         = errors.New("order with this idempotency key already exists")
	ErrInvalidTransition      = errors.New("invalid status transition")
	ErrCancellationExpired    = errors.New("cancellation window has expired")
	// ErrStatusChanged means another request changed the order's status first
	// (e.g. the shop accepted it while the customer was cancelling).
	ErrStatusChanged = errors.New("order status changed, refresh and try again")
)
