package outbox

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/quickeats/pkg/events"
)

// Relay polls the outbox table and publishes unpublished events to the broker.
type Relay struct {
	db        *pgxpool.Pool
	publisher events.Publisher
	interval  time.Duration
	stop      chan struct{}
}

func NewRelay(db *pgxpool.Pool, publisher events.Publisher, interval time.Duration) *Relay {
	return &Relay{
		db:        db,
		publisher: publisher,
		interval:  interval,
		stop:      make(chan struct{}),
	}
}

func (r *Relay) Start(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	slog.Info("outbox relay started", "interval", r.interval)
	for {
		select {
		case <-ticker.C:
			r.pollAndPublish(ctx)
		case <-r.stop:
			slog.Info("outbox relay stopped")
			return
		case <-ctx.Done():
			return
		}
	}
}

func (r *Relay) Stop() {
	close(r.stop)
}

func (r *Relay) pollAndPublish(ctx context.Context) {
	rows, err := r.db.Query(ctx,
		`SELECT id, aggregate_type, aggregate_id, event_type, payload
		 FROM outbox
		 WHERE published_at IS NULL
		 ORDER BY created_at ASC
		 LIMIT 100`)
	if err != nil {
		slog.Error("outbox poll failed", "error", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var aggregateType, aggregateID, eventType string
		var payload json.RawMessage

		if err := rows.Scan(&id, &aggregateType, &aggregateID, &eventType, &payload); err != nil {
			slog.Error("outbox scan failed", "error", err)
			continue
		}

		var event events.Event
		if err := json.Unmarshal(payload, &event); err != nil {
			slog.Error("outbox unmarshal failed", "id", id, "error", err)
			continue
		}

		// Publish to the event topic (e.g., "order.order.placed")
		if err := r.publisher.Publish(ctx, eventType, event); err != nil {
			slog.Error("outbox publish failed", "id", id, "event_type", eventType, "error", err)
			continue
		}

		// Mark as published
		_, err := r.db.Exec(ctx,
			`UPDATE outbox SET published_at = $1 WHERE id = $2`,
			time.Now().UTC(), id)
		if err != nil {
			slog.Error("outbox mark published failed", "id", id, "error", err)
		}
	}
}
