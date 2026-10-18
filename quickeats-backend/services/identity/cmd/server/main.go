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
	"github.com/quickeats/services/identity/internal/adapter/jwt"
	pgRepo "github.com/quickeats/services/identity/internal/adapter/postgres"
	"github.com/quickeats/services/identity/internal/adapter/gateway"
	httpAdapter "github.com/quickeats/services/identity/internal/adapter/http"
	"github.com/quickeats/services/identity/internal/app"
)

func main() {
	cfg := config.Load()

	// Database
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		slog.Error("failed to ping database", "error", err)
		os.Exit(1)
	}
	slog.Info("connected to database")

	// Repositories
	userRepo := pgRepo.NewUserRepo(pool)
	orgRepo := pgRepo.NewOrgRepo(pool)

	// Token issuer
	tokenIssuer := jwt.NewIssuer(cfg.JWTSecret, cfg.JWTExpiry, cfg.JWTRefreshExpiry)

	// Use cases
	registerUC := app.NewRegisterUseCase(userRepo, orgRepo)
	registerShopUC := app.NewRegisterShopUseCase(userRepo, orgRepo, gateway.NewCatalogClient(cfg.CatalogServiceURL))
	loginUC := app.NewLoginUseCase(userRepo, tokenIssuer)

	// HTTP handlers
	authHandler := httpAdapter.NewAuthHandler(registerUC, registerShopUC, loginUC, tokenIssuer, userRepo)
	orgHandler := httpAdapter.NewOrgHandler(orgRepo)

	// Router
	router := httpAdapter.NewRouter(authHandler, orgHandler, cfg.JWTSecret)

	// Server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Graceful shutdown
	go func() {
		slog.Info("identity service starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
