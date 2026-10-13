package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

type NATSPublisher struct {
	js nats.JetStreamContext
	nc *nats.Conn
}

func NewNATSPublisher(url string) (*NATSPublisher, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("jetstream: %w", err)
	}

	// Ensure streams exist
	for _, stream := range []struct {
		Name     string
		Subjects []string
	}{
		{"ORDER", []string{"order.>"}},
		{"PAYMENT", []string{"payment.>"}},
		{"CATALOG", []string{"catalog.>"}},
	} {
		_, err := js.AddStream(&nats.StreamConfig{
			Name:     stream.Name,
			Subjects: stream.Subjects,
			Storage:  nats.FileStorage,
		})
		if err != nil {
			// Stream may already exist, ignore
		}
	}

	return &NATSPublisher{js: js, nc: nc}, nil
}

func (p *NATSPublisher) Publish(ctx context.Context, subject string, event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	_, err = p.js.Publish(subject, data)
	if err != nil {
		return fmt.Errorf("publish to %s: %w", subject, err)
	}
	return nil
}

func (p *NATSPublisher) Close() error {
	p.nc.Close()
	return nil
}
