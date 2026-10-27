package events

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	topicPartitions  = 3
	topicReplication = 1 // single broker
)

// ParseBrokers splits a comma-separated broker list ("kafka:29092,other:29092").
func ParseBrokers(list string) []string {
	var brokers []string
	for _, b := range strings.Split(list, ",") {
		if b = strings.TrimSpace(b); b != "" {
			brokers = append(brokers, b)
		}
	}
	return brokers
}

// waitForBroker retries until a broker accepts connections. Kafka takes a few
// seconds to start, so services that boot first wait instead of crashing.
func waitForBroker(brokers []string) error {
	var lastErr error
	for attempt := 0; attempt < 30; attempt++ {
		for _, b := range brokers {
			conn, err := kafka.Dial("tcp", b)
			if err == nil {
				conn.Close()
				return nil
			}
			lastErr = err
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("kafka unreachable at %v: %w", brokers, lastErr)
}

// waitForCoordinator retries until the broker can serve the consumer group. On a
// fresh broker the internal offsets topic is created by the first group request, and
// readers that try to join before it is ready can get stuck outside their group.
func waitForCoordinator(brokers []string, groupID string) error {
	client := &kafka.Client{Addr: kafka.TCP(brokers...), Timeout: 5 * time.Second}
	var lastErr error
	for attempt := 0; attempt < 30; attempt++ {
		resp, err := client.FindCoordinator(context.Background(), &kafka.FindCoordinatorRequest{
			Key:     groupID,
			KeyType: kafka.CoordinatorKeyTypeConsumer,
		})
		if err == nil && resp.Error == nil {
			return nil
		}
		if err == nil {
			err = resp.Error
		}
		lastErr = err
		time.Sleep(time.Second)
	}
	return fmt.Errorf("no coordinator for group %s: %w", groupID, lastErr)
}

// topicManager creates topics on first use. Both publishers and consumers call
// it, so start-up order between services doesn't matter.
type topicManager struct {
	brokers []string
	created sync.Map
}

func (m *topicManager) ensure(topic string) error {
	if _, ok := m.created.Load(topic); ok {
		return nil
	}

	conn, err := kafka.Dial("tcp", m.brokers[0])
	if err != nil {
		return fmt.Errorf("dial broker: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("find controller: %w", err)
	}
	ctrl, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return fmt.Errorf("dial controller: %w", err)
	}
	defer ctrl.Close()

	err = ctrl.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     topicPartitions,
		ReplicationFactor: topicReplication,
	})
	if err != nil && !errors.Is(err, kafka.TopicAlreadyExists) {
		return fmt.Errorf("create topic %s: %w", topic, err)
	}

	m.created.Store(topic, true)
	slog.Info("kafka topic ready", "topic", topic)
	return nil
}
