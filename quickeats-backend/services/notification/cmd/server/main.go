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
	"github.com/quickeats/pkg/config"
	"github.com/quickeats/pkg/events"
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

	consumer.Subscribe("order.order.confirmed", "notification-svc", notifService.HandleOrderConfirmed)
	consumer.Subscribe("order.order.cancelled", "notification-svc", notifService.HandleOrderCancelled)
	consumer.Subscribe("order.order.rejected", "notification-svc", notifService.HandleOrderRejected)
	consumer.Subscribe("order.order.status_changed", "notification-svc", notifService.HandleOrderStatusChanged)

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
