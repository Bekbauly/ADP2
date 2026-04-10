package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"order-service/internal/repository"
	transporthttp "order-service/internal/transport/http"
	"order-service/internal/usecase"
)

type Config struct {
	HTTPPort          string
	DatabaseDSN       string
	PaymentServiceURL string
}
type App struct {
	cfg    Config
	db     *sql.DB
	router *gin.Engine
}

func New(cfg Config) (*App, error) {
	db, err := sql.Open("postgres", cfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	log.Println("Connected to Order DB")

	orderRepo := repository.NewPostgresOrderRepository(db)
	httpClient := &http.Client{Timeout: 2 * time.Second}
	paymentClient := transporthttp.NewPaymentHTTPClient(httpClient, cfg.PaymentServiceURL)
	orderUC := usecase.NewOrderUseCase(orderRepo, paymentClient)
	orderHandler := transporthttp.NewOrderHandler(orderUC)

	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "order", "status": "ok"})
	})
	v1 := router.Group("/api/v1")
	orderHandler.RegisterRoutes(v1)

	return &App{cfg: cfg, db: db, router: router}, nil
}

func (a *App) Run() error {
	log.Printf("Order Service listening on :%s", a.cfg.HTTPPort)
	return a.router.Run(":" + a.cfg.HTTPPort)
}

func (a *App) Close() {
	if a.db != nil {
		a.db.Close()
	}
}
