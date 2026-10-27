package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

// maxHandlerAttempts bounds retries of a failing handler. Kafka has no per-message
// nack: an uncommitted message blocks its partition, so after the last attempt the
// event is logged and skipped rather than stalling every event behind it.
const maxHandlerAttempts = 5

type KafkaConsumer struct {
	brokers []string
	topics  *topicManager
	ctx     context.Context
	cancel  context.CancelFunc
	readers []*kafka.Reader
	wg      sync.WaitGroup
}

func NewKafkaConsumer(brokers []string) (*KafkaConsumer, error) {
	if err := waitForBroker(brokers); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &KafkaConsumer{
		brokers: brokers,
		topics:  &topicManager{brokers: brokers},
		ctx:     ctx,
		cancel:  cancel,
	}, nil
}

func (c *KafkaConsumer) Subscribe(subject string, group string, handler Handler) error {
	if err := c.topics.ensure(subject); err != nil {
		slog.Error("event subscription failed", "subject", subject, "error", err)
		return fmt.Errorf("subscribe %s: %w", subject, err)
	}

	// One consumer group per subject, so every topic keeps its own committed offset.
	groupID := group + "_" + strings.ReplaceAll(subject, ".", "_")
	if err := waitForCoordinator(c.brokers, groupID); err != nil {
		slog.Error("event subscription failed", "subject", subject, "error", err)
		return fmt.Errorf("subscribe %s: %w", subject, err)
	}
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     c.brokers,
		GroupID:     groupID,
		Topic:       subject,
		StartOffset: kafka.FirstOffset, // a new group also receives events published before it existed
		MaxWait:     500 * time.Millisecond,
		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			slog.Error("kafka reader: "+fmt.Sprintf(msg, args...), "subject", subject)
		}),
	})
	c.readers = append(c.readers, reader)

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.consume(reader, subject, handler)
	}()
	return nil
}

func (c *KafkaConsumer) consume(reader *kafka.Reader, subject string, handler Handler) {
	for {
		msg, err := reader.FetchMessage(c.ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) { // shutting down
				return
			}
			slog.Error("kafka fetch failed", "subject", subject, "error", err)
			time.Sleep(time.Second)
			continue
		}

		c.handle(subject, msg, handler)

		// Manual commit after handling: a crash before this line redelivers the event.
		if err := reader.CommitMessages(c.ctx, msg); err != nil && c.ctx.Err() == nil {
			slog.Error("kafka commit failed", "subject", subject, "offset", msg.Offset, "error", err)
		}
	}
}

func (c *KafkaConsumer) handle(subject string, msg kafka.Message, handler Handler) {
	var event Event
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		slog.Error("failed to unmarshal event, skipping", "subject", subject, "offset", msg.Offset, "error", err)
		return
	}

	backoff := 500 * time.Millisecond
	for attempt := 1; attempt <= maxHandlerAttempts; attempt++ {
		err := handler(c.ctx, event)
		if err == nil {
			return
		}
		if attempt == maxHandlerAttempts || c.ctx.Err() != nil {
			slog.Error("failed to handle event, giving up", "event_type", event.EventType, "event_id", event.EventID, "attempts", attempt, "error", err)
			return
		}
		slog.Warn("failed to handle event, retrying", "event_type", event.EventType, "attempt", attempt, "error", err)
		time.Sleep(backoff)
		backoff *= 2
	}
}

func (c *KafkaConsumer) Close() error {
	c.cancel()
	for _, r := range c.readers {
		r.Close()
	}
	c.wg.Wait()
	return nil
}
