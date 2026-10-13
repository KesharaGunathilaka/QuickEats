package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/quickeats/pkg/events"
)

// Entry represents a row in the outbox table.
type Entry struct {
	ID            int64
	AggregateType string
	AggregateID   string
	EventType     string
	Payload       json.RawMessage
	CreatedAt     time.Time
	PublishedAt   *time.Time
}

// WriteToOutbox inserts an event into the outbox table within an existing transaction.
// This guarantees the event is saved atomically with the business data change.
func WriteToOutbox(ctx context.Context, tx pgx.Tx, event events.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event payload: %w", err)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO outbox (aggregate_type, aggregate_id, event_type, payload, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		event.AggregateType,
		event.AggregateID,
		event.EventType,
		payload,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("insert outbox: %w", err)
	}
	return nil
}
