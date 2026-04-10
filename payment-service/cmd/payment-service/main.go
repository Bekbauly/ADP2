package main

import (
	"log"
	"os"

	"payment-service/internal/app"
)

func main() {
	cfg := app.Config{
		HTTPPort:    getEnv("HTTP_PORT", "8081"),
		DatabaseDSN: getEnv("DATABASE_DSN", "host=localhost port=5432 user=postgres password=postgres dbname=payments_db sslmode=disable"),
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to initialize payment service: %v", err)
	}
	defer application.Close()

	if err := application.Run(); err != nil {
		log.Fatalf("payment service error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
