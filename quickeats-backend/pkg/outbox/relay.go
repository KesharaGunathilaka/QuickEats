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

// pollAndPublish publishes the oldest unpublished outbox rows and marks them published.
//
// Delivery is at-least-once: publishing to Kafka and marking the row cannot be one
// atomic step, so a crash after publishing but before commit re-publishes those rows
// on the next poll. That is the accepted trade-off of the outbox pattern; each copy
// carries the same EventID, and consumers skip copies with events.Deduplicate.
//
// The rows are locked (FOR UPDATE SKIP LOCKED) for the whole batch, so several
// replicas of a service can run relays without publishing the same row twice.
func (r *Relay) pollAndPublish(ctx context.Context) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		slog.Error("outbox poll: begin tx failed", "error", err)
		return
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx,
		`SELECT id, event_type, payload
		 FROM outbox
		 WHERE published_at IS NULL
		 ORDER BY created_at ASC, id ASC
		 LIMIT 100
		 FOR UPDATE SKIP LOCKED`)
	if err != nil {
		slog.Error("outbox poll failed", "error", err)
		return
	}

	type pending struct {
		id        int64
		eventType string
		payload   json.RawMessage
	}
	var batch []pending
	for rows.Next() {
		var p pending
		if err := rows.Scan(&p.id, &p.eventType, &p.payload); err != nil {
			slog.Error("outbox scan failed", "error", err)
			continue
		}
		batch = append(batch, p)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		slog.Error("outbox poll failed", "error", err)
		return
	}

	published := 0
	for _, p := range batch {
		var event events.Event
		if err := json.Unmarshal(p.payload, &event); err != nil {
			// A row that can never be decoded would block the queue forever: mark it
			// so the relay moves on, and log it for investigation.
			slog.Error("outbox unmarshal failed, skipping row", "id", p.id, "error", err)
		} else if err := r.publisher.Publish(ctx, p.eventType, event); err != nil {
			// Stop at the first failure so later events don't overtake this one;
			// it is retried first on the next poll.
			slog.Error("outbox publish failed", "id", p.id, "event_type", p.eventType, "error", err)
			break
		}

		if _, err := tx.Exec(ctx, `UPDATE outbox SET published_at = $1 WHERE id = $2`, time.Now().UTC(), p.id); err != nil {
			slog.Error("outbox mark published failed", "id", p.id, "error", err)
			break
		}
		published++
	}

	if published == 0 {
		return
	}
	if err := tx.Commit(ctx); err != nil {
		// The events went out but stay unmarked, so they are re-published next poll;
		// consumers drop the duplicates.
		slog.Error("outbox commit failed, events will be re-published", "count", published, "error", err)
	}
}
