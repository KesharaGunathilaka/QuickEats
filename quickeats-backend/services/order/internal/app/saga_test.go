package app

import (
	"context"
	"errors"
	"testing"

	"github.com/quickeats/pkg/events"
	"github.com/quickeats/services/order/internal/domain"
	"github.com/quickeats/services/order/internal/port"
)

// fakeOrderRepo implements only GetByID; the embedded interface panics if the
// saga calls anything else.
type fakeOrderRepo struct {
	port.OrderRepository
	orders map[string]*domain.Order
	err    error
	calls  int
}

func (f *fakeOrderRepo) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return f.orders[id], nil
}

// Payloads arrive from Kafka as decoded JSON, i.e. map[string]interface{}.
func sagaEvent(payload interface{}) events.Event {
	return events.Event{EventID: "evt-1", EventType: "payment.payment.failed", Payload: payload}
}

func TestLoadOrder(t *testing.T) {
	existing := &domain.Order{ID: "order-1", Status: domain.StatusPendingPayment}
	dbDown := errors.New("connection refused")

	tests := []struct {
		name         string
		payload      interface{}
		repoErr      error
		wantOrder    bool
		wantErr      bool
		wantRepoCall bool
	}{
		{name: "valid payload loads the order", payload: map[string]interface{}{"order_id": "order-1"}, wantOrder: true, wantRepoCall: true},
		{name: "missing order_id is skipped", payload: map[string]interface{}{"other": "x"}},
		{name: "empty order_id is skipped", payload: map[string]interface{}{"order_id": ""}},
		{name: "wrong type for order_id is skipped", payload: map[string]interface{}{"order_id": 42}},
		{name: "payload that is not an object is skipped", payload: "not-an-object"},
		{name: "nil payload is skipped", payload: nil},
		{name: "unknown order is skipped", payload: map[string]interface{}{"order_id": "missing"}, wantRepoCall: true},
		{name: "database error is returned for retry", payload: map[string]interface{}{"order_id": "order-1"}, repoErr: dbDown, wantErr: true, wantRepoCall: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeOrderRepo{orders: map[string]*domain.Order{"order-1": existing}, err: tt.repoErr}
			saga := NewSagaOrchestrator(nil, repo, nil, nil)

			order, err := saga.loadOrder(context.Background(), sagaEvent(tt.payload))

			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !errors.Is(err, dbDown) {
				t.Errorf("err = %v, want it to wrap the database error", err)
			}
			if (order != nil) != tt.wantOrder {
				t.Errorf("order = %v, wantOrder %v", order, tt.wantOrder)
			}
			if (repo.calls > 0) != tt.wantRepoCall {
				t.Errorf("repo calls = %d, wantRepoCall %v", repo.calls, tt.wantRepoCall)
			}
		})
	}
}

// A malformed event must not reach compensation (which would touch the DB and
// catalog, both nil here) and must not return an error that makes the consumer retry.
func TestHandlersSkipMalformedEvents(t *testing.T) {
	handlers := map[string]func(*SagaOrchestrator) func(context.Context, events.Event) error{
		"HandleOrderPlaced":       func(s *SagaOrchestrator) func(context.Context, events.Event) error { return s.HandleOrderPlaced },
		"HandleStockReserved":     func(s *SagaOrchestrator) func(context.Context, events.Event) error { return s.HandleStockReserved },
		"HandlePaymentCompleted":  func(s *SagaOrchestrator) func(context.Context, events.Event) error { return s.HandlePaymentCompleted },
		"HandlePaymentFailed":     func(s *SagaOrchestrator) func(context.Context, events.Event) error { return s.HandlePaymentFailed },
		"HandleStockInsufficient": func(s *SagaOrchestrator) func(context.Context, events.Event) error { return s.HandleStockInsufficient },
	}

	for name, handler := range handlers {
		t.Run(name, func(t *testing.T) {
			repo := &fakeOrderRepo{}
			saga := NewSagaOrchestrator(nil, repo, nil, nil)

			err := handler(saga)(context.Background(), sagaEvent(map[string]interface{}{"order_id": 42}))

			if err != nil {
				t.Errorf("err = %v, want nil (skip, not retry)", err)
			}
			if repo.calls != 0 {
				t.Errorf("repo called %d times for a malformed event", repo.calls)
			}
		})
	}
}
