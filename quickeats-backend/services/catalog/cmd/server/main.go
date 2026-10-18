package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	redisLib "github.com/redis/go-redis/v9"
	"github.com/quickeats/pkg/config"
	httpAdapter "github.com/quickeats/services/catalog/internal/adapter/http"
	pgRepo "github.com/quickeats/services/catalog/internal/adapter/postgres"
	redisAdapter "github.com/quickeats/services/catalog/internal/adapter/redis"
	"github.com/quickeats/services/catalog/internal/app"
)

func main() {
	cfg := config.Load()

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil { slog.Error("db connect failed", "error", err); os.Exit(1) }
	defer pool.Close()
	pool.Ping(context.Background())
	slog.Info("connected to database")

	rdb := redisLib.NewClient(&redisLib.Options{Addr: cfg.RedisURL})
	defer rdb.Close()
	slog.Info("connected to redis")
	warmStockCache(context.Background(), pool, rdb)

	shopRepo := pgRepo.NewShopRepo(pool)
	menuRepo := pgRepo.NewMenuItemRepo(pool)
	stockCache := redisAdapter.NewStockCache(rdb)

	browseUC := app.NewBrowseShopsUseCase(shopRepo)
	menuUC := app.NewManageMenuUseCase(menuRepo, stockCache)
	stockUC := app.NewReserveStockUseCase(pool, stockCache)

	handler := httpAdapter.NewCatalogHandler(browseUC, menuUC, stockUC, shopRepo)
	router := httpAdapter.NewRouter(handler, cfg.JWTSecret)

	srv := &http.Server{Addr: fmt.Sprintf(":%s", cfg.Port), Handler: router, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second}

	go func() {
		slog.Info("catalog service starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed { slog.Error("server error", "error", err); os.Exit(1) }
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

// warmStockCache loads stock counts for items that have no Redis key yet
// (e.g. seeded rows), so reservations don't treat them as out of stock.
func warmStockCache(ctx context.Context, pool *pgxpool.Pool, rdb *redisLib.Client) {
	rows, err := pool.Query(ctx, `SELECT id, stock FROM menu_items`)
	if err != nil { slog.Error("stock cache warm-up failed", "error", err); return }
	defer rows.Close()
	warmed := 0
	for rows.Next() {
		var id string
		var stock int
		if err := rows.Scan(&id, &stock); err != nil { continue }
		if ok, _ := rdb.SetNX(ctx, fmt.Sprintf("stock:%s", id), stock, 0).Result(); ok { warmed++ }
	}
	slog.Info("stock cache warmed", "items", warmed)
}
