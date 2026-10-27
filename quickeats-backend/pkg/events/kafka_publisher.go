package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	writer *kafka.Writer
	topics *topicManager
}

func NewKafkaPublisher(brokers []string) (*KafkaPublisher, error) {
	if err := waitForBroker(brokers); err != nil {
		return nil, err
	}

	return &KafkaPublisher{
		writer: &kafka.Writer{
			Addr: kafka.TCP(brokers...),
			// Same key -> same partition, so events of one aggregate (e.g. one
			// order) are consumed in the order they were published.
			Balancer:     &kafka.Hash{},
			RequiredAcks: kafka.RequireAll,
		},
		topics: &topicManager{brokers: brokers},
	}, nil
}

// Publish sends the event to the topic named by subject (e.g. "order.order.placed").
func (p *KafkaPublisher) Publish(ctx context.Context, subject string, event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	if err := p.topics.ensure(subject); err != nil {
		return err
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Topic: subject,
		Key:   []byte(event.AggregateID),
		Value: data,
	})
	if err != nil {
		return fmt.Errorf("publish to %s: %w", subject, err)
	}
	return nil
}

func (p *KafkaPublisher) Close() error {
	return p.writer.Close()
}
