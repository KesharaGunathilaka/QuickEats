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
