package events

import (
	"context"
	"log/slog"
	"time"
)

// ProcessedStore remembers which events a consumer has already handled.
type ProcessedStore interface {
	IsProcessed(ctx context.Context, key string) (bool, error)
	MarkProcessed(ctx context.Context, key string, ttl time.Duration) error
}

// processedTTL is how long a handled event ID is remembered. Duplicates come from
// the outbox relay re-publishing after a crash, or Kafka redelivering an uncommitted
// message, so they arrive within minutes; a week leaves a wide margin.
const processedTTL = 7 * 24 * time.Hour

// Deduplicate wraps a handler so each event is handled at most once per consumer.
//
// Delivery is at-least-once: the outbox relay cannot publish to Kafka and mark the
// row published atomically, and the consumer commits offsets only after handling.
// Every copy of an event carries the same EventID (it is stored in the outbox row),
// so this wrapper skips copies it has already handled.
//
// The event is marked only after the handler succeeds, so a failed or interrupted
// attempt is retried. A crash between handling and marking can still let one
// duplicate through, which is why handlers also guard on business state (order
// status, one payment per order, one reservation set per order).
func Deduplicate(store ProcessedStore, consumer string, handler Handler) Handler {
	return func(ctx context.Context, event Event) error {
		if event.EventID == "" {
			return handler(ctx, event)
		}
		key := consumer + ":" + event.EventID

		seen, err := store.IsProcessed(ctx, key)
		if err != nil {
			// Store unavailable: handle anyway and rely on the handler's own guards,
			// rather than stalling the topic.
			slog.Warn("dedupe check failed, handling anyway", "event_id", event.EventID, "error", err)
		} else if seen {
			slog.Info("duplicate event skipped", "consumer", consumer, "event_type", event.EventType, "event_id", event.EventID)
			return nil
		}

		if err := handler(ctx, event); err != nil {
			return err
		}

		if err := store.MarkProcessed(ctx, key, processedTTL); err != nil {
			slog.Warn("dedupe mark failed", "event_id", event.EventID, "error", err)
		}
		return nil
	}
}
