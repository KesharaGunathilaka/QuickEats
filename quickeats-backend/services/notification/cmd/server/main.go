package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	redisLib "github.com/redis/go-redis/v9"
	"github.com/quickeats/pkg/config"
	"github.com/quickeats/pkg/events"
	"github.com/quickeats/pkg/idempotency"
	"github.com/quickeats/services/notification/internal/app"
	httpAdapter "github.com/quickeats/services/notification/internal/adapter/http"
)

func main() {
	cfg := config.Load()

	hub := app.NewHub()
	notifService := app.NewNotificationService(hub)

	// Event consumer
	consumer, err := events.NewKafkaConsumer(events.ParseBrokers(cfg.KafkaBrokers))
	if err != nil { slog.Error("kafka failed", "error", err); os.Exit(1) }
	defer consumer.Close()

	// Delivery is at-least-once; skip events already pushed so users aren't notified twice.
	rdb := redisLib.NewClient(&redisLib.Options{Addr: cfg.RedisURL})
	defer rdb.Close()
	processed := idempotency.NewChecker(rdb)
	once := func(h events.Handler) events.Handler { return events.Deduplicate(processed, "notification-svc", h) }

	consumer.Subscribe("order.order.placed", "notification-svc", once(notifService.HandleOrderPlaced))
	consumer.Subscribe("order.order.confirmed", "notification-svc", once(notifService.HandleOrderConfirmed))
	consumer.Subscribe("order.order.cancelled", "notification-svc", once(notifService.HandleOrderCancelled))
	consumer.Subscribe("order.order.rejected", "notification-svc", once(notifService.HandleOrderRejected))
	consumer.Subscribe("order.order.status_changed", "notification-svc", once(notifService.HandleOrderStatusChanged))

	// HTTP health check
	httpRouter := chi.NewRouter()
	httpRouter.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok", "service": "notification", "activeConnections": hub.ActiveConnections()})
	})
	go func() {
		slog.Info("notification HTTP starting", "port", cfg.HTTPPort)
		http.ListenAndServe(fmt.Sprintf(":%s", cfg.HTTPPort), httpRouter)
	}()

	// WebSocket server
	wsHandler := httpAdapter.NewWSHandler(hub, cfg.JWTSecret)
	wsMux := http.NewServeMux()
	wsMux.HandleFunc("/ws", wsHandler.HandleWS)
	go func() {
		slog.Info("notification WebSocket starting", "port", cfg.WSPort)
		http.ListenAndServe(fmt.Sprintf(":%s", cfg.WSPort), wsMux)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("notification service stopping")
}
