package events

import "context"

// Handler processes a received event.
type Handler func(ctx context.Context, event Event) error

// Consumer defines the interface for subscribing to events.
type Consumer interface {
	Subscribe(subject string, group string, handler Handler) error
	Close() error
}
