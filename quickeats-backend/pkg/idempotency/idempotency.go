package idempotency

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Checker struct {
	rdb *redis.Client
}

func NewChecker(rdb *redis.Client) *Checker {
	return &Checker{rdb: rdb}
}

// Check returns true if the key is new (first time seen), false if duplicate.
func (c *Checker) Check(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	set, err := c.rdb.SetNX(ctx, fmt.Sprintf("idempotency:%s", key), "1", ttl).Result()
	if err != nil {
		return false, fmt.Errorf("idempotency check: %w", err)
	}
	return set, nil
}

// IsProcessed reports whether a consumer already handled an event (see events.Deduplicate).
func (c *Checker) IsProcessed(ctx context.Context, key string) (bool, error) {
	n, err := c.rdb.Exists(ctx, fmt.Sprintf("processed:%s", key)).Result()
	if err != nil {
		return false, fmt.Errorf("processed check: %w", err)
	}
	return n > 0, nil
}

// MarkProcessed records that a consumer handled an event.
func (c *Checker) MarkProcessed(ctx context.Context, key string, ttl time.Duration) error {
	if err := c.rdb.Set(ctx, fmt.Sprintf("processed:%s", key), "1", ttl).Err(); err != nil {
		return fmt.Errorf("mark processed: %w", err)
	}
	return nil
}
