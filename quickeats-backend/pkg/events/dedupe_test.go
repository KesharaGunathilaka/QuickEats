package events

import (
	"context"
	"errors"
	"testing"
	"time"
)

type memStore struct {
	seen      map[string]bool
	checkErr  error
	markCalls int
}

func newMemStore() *memStore { return &memStore{seen: map[string]bool{}} }

func (m *memStore) IsProcessed(ctx context.Context, key string) (bool, error) {
	if m.checkErr != nil {
		return false, m.checkErr
	}
	return m.seen[key], nil
}

func (m *memStore) MarkProcessed(ctx context.Context, key string, ttl time.Duration) error {
	m.markCalls++
	m.seen[key] = true
	return nil
}

// countingHandler counts calls and fails while failures > 0.
func countingHandler(calls *int, failures int) Handler {
	return func(ctx context.Context, event Event) error {
		*calls++
		if failures > 0 {
			failures--
			return errors.New("transient failure")
		}
		return nil
	}
}

func TestDeduplicateSkipsRepublishedEvent(t *testing.T) {
	store := newMemStore()
	calls := 0
	h := Deduplicate(store, "order-svc", countingHandler(&calls, 0))
	event := Event{EventID: "evt-1", EventType: "order.order.placed"}

	// The relay re-publishes the same outbox row: same EventID both times.
	for i := 0; i < 3; i++ {
		if err := h(context.Background(), event); err != nil {
			t.Fatalf("delivery %d: unexpected error %v", i+1, err)
		}
	}
	if calls != 1 {
		t.Errorf("handler ran %d times for one event, want 1", calls)
	}
}

func TestDeduplicateRetriesAfterFailure(t *testing.T) {
	store := newMemStore()
	calls := 0
	h := Deduplicate(store, "order-svc", countingHandler(&calls, 1))
	event := Event{EventID: "evt-1"}

	if err := h(context.Background(), event); err == nil {
		t.Fatal("first attempt: want the handler's error, got nil")
	}
	if store.seen["order-svc:evt-1"] {
		t.Fatal("a failed attempt must not be marked processed")
	}
	if err := h(context.Background(), event); err != nil {
		t.Fatalf("retry: unexpected error %v", err)
	}
	if calls != 2 {
		t.Errorf("handler ran %d times, want 2 (fail, then succeed)", calls)
	}
}

func TestDeduplicateIsPerConsumer(t *testing.T) {
	store := newMemStore()
	orderCalls, notifCalls := 0, 0
	orderH := Deduplicate(store, "order-svc", countingHandler(&orderCalls, 0))
	notifH := Deduplicate(store, "notification-svc", countingHandler(&notifCalls, 0))
	event := Event{EventID: "evt-1"}

	orderH(context.Background(), event)
	notifH(context.Background(), event)

	if orderCalls != 1 || notifCalls != 1 {
		t.Errorf("order=%d notification=%d, want each consumer to handle the event once", orderCalls, notifCalls)
	}
}

func TestDeduplicateHandlesWhenStoreIsDown(t *testing.T) {
	store := newMemStore()
	store.checkErr = errors.New("redis unavailable")
	calls := 0
	h := Deduplicate(store, "order-svc", countingHandler(&calls, 0))

	if err := h(context.Background(), Event{EventID: "evt-1"}); err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if calls != 1 {
		t.Errorf("handler ran %d times, want 1 (fall back to handling)", calls)
	}
}

func TestDeduplicatePassesThroughEventsWithoutID(t *testing.T) {
	store := newMemStore()
	calls := 0
	h := Deduplicate(store, "order-svc", countingHandler(&calls, 0))

	h(context.Background(), Event{})
	h(context.Background(), Event{})

	if calls != 2 {
		t.Errorf("handler ran %d times, want 2 (nothing to dedupe on)", calls)
	}
	if store.markCalls != 0 {
		t.Errorf("marked %d events without an ID, want 0", store.markCalls)
	}
}
