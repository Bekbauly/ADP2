package main

import (
	"log"
	"os"

	"order-service/internal/app"
)

func main() {
	cfg := app.Config{
		HTTPPort:          getEnv("HTTP_PORT", "8080"),
		DatabaseDSN:       getEnv("DATABASE_DSN", "host=localhost port=5432 user=postgres password=postgres dbname=orders_db sslmode=disable"),
		PaymentServiceURL: getEnv("PAYMENT_SERVICE_URL", "http://localhost:8081/api/v1"),
	}
	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to initialize order service: %v", err)
	}
	defer application.Close()

	if err := application.Run(); err != nil {
		log.Fatalf("order service error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
