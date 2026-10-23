package config

import (
	"os"
	"time"
)

type Config struct {
	Port              string
	DatabaseURL       string
	RedisURL          string
	NatsURL           string
	JWTSecret         string
	JWTExpiry         time.Duration
	JWTRefreshExpiry  time.Duration
	CatalogServiceURL string
	PaymentServiceURL string
	OrderServiceURL   string
	HTTPPort          string
	WSPort            string
}

func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "8080"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://quickeats:quickeats@localhost:5433/quickeats_db?sslmode=disable"),
		RedisURL:          getEnv("REDIS_URL", "localhost:6379"),
		NatsURL:           getEnv("NATS_URL", "nats://localhost:4222"),
		JWTSecret:         getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		JWTExpiry:         parseDuration(getEnv("JWT_EXPIRY", "15m")),
		JWTRefreshExpiry:  parseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h")),
		CatalogServiceURL: getEnv("CATALOG_SERVICE_URL", "http://localhost:8082"),
		PaymentServiceURL: getEnv("PAYMENT_SERVICE_URL", "http://localhost:8084"),
		OrderServiceURL:   getEnv("ORDER_SERVICE_URL", "http://localhost:8083"),
		HTTPPort:          getEnv("HTTP_PORT", "8085"),
		WSPort:            getEnv("WS_PORT", "8086"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 15 * time.Minute
	}
	return d
}
