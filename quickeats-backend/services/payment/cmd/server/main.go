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
	"github.com/quickeats/pkg/config"
	"github.com/quickeats/pkg/events"
	"github.com/quickeats/pkg/outbox"
	"github.com/quickeats/services/payment/internal/adapter/gateway"
	httpAdapter "github.com/quickeats/services/payment/internal/adapter/http"
	pgRepo "github.com/quickeats/services/payment/internal/adapter/postgres"
	"github.com/quickeats/services/payment/internal/app"
	"github.com/quickeats/services/payment/internal/domain"
)

func main() {
	cfg := config.Load()

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil { slog.Error("db failed", "error", err); os.Exit(1) }
	defer pool.Close()
	pool.Ping(context.Background())

	publisher, err := events.NewKafkaPublisher(events.ParseBrokers(cfg.KafkaBrokers))
	if err != nil { slog.Error("kafka failed", "error", err); os.Exit(1) }
	defer publisher.Close()

	paymentRepo := pgRepo.NewPaymentRepo(pool)
	ledgerRepo := pgRepo.NewLedgerRepo(pool)
	cardRepo := pgRepo.NewCardRepo(pool)

	// Card is the only payment method: the customer's saved card is charged directly.
	strategies := map[string]domain.PaymentStrategy{
		"card": gateway.NewCardStrategy(cardRepo),
	}

	paymentUC := app.NewCreatePaymentUseCase(pool, paymentRepo, ledgerRepo, strategies)
	cardUC := app.NewCardUseCase(cardRepo)

	relay := outbox.NewRelay(pool, publisher, 2*time.Second)
	go relay.Start(context.Background())

	router := httpAdapter.NewRouter(paymentUC, cardUC, cfg.JWTSecret)

	srv := &http.Server{Addr: fmt.Sprintf(":%s", cfg.Port), Handler: router, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second}
	go func() {
		slog.Info("payment service starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed { slog.Error("error", "error", err); os.Exit(1) }
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	relay.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
