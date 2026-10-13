package domain

import "errors"

var (
	ErrOrderNotFound          = errors.New("order not found")
	ErrDuplicateOrder         = errors.New("order with this idempotency key already exists")
	ErrInvalidTransition      = errors.New("invalid status transition")
	ErrCancellationExpired    = errors.New("cancellation window has expired")
)
