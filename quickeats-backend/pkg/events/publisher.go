package events

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Event is the standard event envelope used across all services.
type Event struct {
	EventID       string      `json:"event_id"`
	EventType     string      `json:"event_type"`
	AggregateType string      `json:"aggregate_type"`
	AggregateID   string      `json:"aggregate_id"`
	Timestamp     time.Time   `json:"timestamp"`
	Version       int         `json:"version"`
	Payload       interface{} `json:"payload"`
}

func NewEvent(eventType, aggregateType, aggregateID string, payload interface{}) Event {
	return Event{
		EventID:       uuid.New().String(),
		EventType:     eventType,
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		Timestamp:     time.Now().UTC(),
		Version:       1,
		Payload:       payload,
	}
}

// Publisher defines the interface for publishing events to the message broker.
type Publisher interface {
	Publish(ctx context.Context, subject string, event Event) error
	Close() error
}
