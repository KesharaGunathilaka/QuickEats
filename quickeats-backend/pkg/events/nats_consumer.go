package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
)

type NATSConsumer struct {
	js   nats.JetStreamContext
	nc   *nats.Conn
	subs []*nats.Subscription
}

func NewNATSConsumer(url string) (*NATSConsumer, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("jetstream: %w", err)
	}
	return &NATSConsumer{js: js, nc: nc}, nil
}

func (c *NATSConsumer) Subscribe(subject string, group string, handler Handler) error {
	sub, err := c.js.QueueSubscribe(subject, group, func(msg *nats.Msg) {
		var event Event
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			slog.Error("failed to unmarshal event", "subject", subject, "error", err)
			msg.Nak()
			return
		}

		if err := handler(context.Background(), event); err != nil {
			slog.Error("failed to handle event", "event_type", event.EventType, "error", err)
			msg.Nak()
			return
		}

		msg.Ack()
	}, nats.Durable(group), nats.ManualAck())
	if err != nil {
		return fmt.Errorf("subscribe %s: %w", subject, err)
	}
	c.subs = append(c.subs, sub)
	return nil
}

func (c *NATSConsumer) Close() error {
	for _, sub := range c.subs {
		sub.Unsubscribe()
	}
	c.nc.Close()
	return nil
}
