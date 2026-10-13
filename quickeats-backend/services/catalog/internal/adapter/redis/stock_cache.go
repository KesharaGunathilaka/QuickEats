package redis

import (
	"context"
	"fmt"

	redisLib "github.com/redis/go-redis/v9"
)

type StockCache struct {
	rdb *redisLib.Client
}

func NewStockCache(rdb *redisLib.Client) *StockCache {
	return &StockCache{rdb: rdb}
}

func (c *StockCache) DecrementStock(ctx context.Context, menuItemID string, quantity int) (int64, error) {
	key := fmt.Sprintf("stock:%s", menuItemID)
	return c.rdb.DecrBy(ctx, key, int64(quantity)).Result()
}

func (c *StockCache) IncrementStock(ctx context.Context, menuItemID string, quantity int) error {
	key := fmt.Sprintf("stock:%s", menuItemID)
	return c.rdb.IncrBy(ctx, key, int64(quantity)).Err()
}

func (c *StockCache) GetStock(ctx context.Context, menuItemID string) (int64, error) {
	key := fmt.Sprintf("stock:%s", menuItemID)
	return c.rdb.Get(ctx, key).Int64()
}

func (c *StockCache) SetStock(ctx context.Context, menuItemID string, stock int) error {
	key := fmt.Sprintf("stock:%s", menuItemID)
	return c.rdb.Set(ctx, key, stock, 0).Err()
}
