package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"order-service/internal/domain"
	"order-service/internal/usecase"
)

type OrderHandler struct {
	uc *usecase.OrderUseCase
}

// NewOrderHandler constructs the handler with the use case injected.
func NewOrderHandler(uc *usecase.OrderUseCase) *OrderHandler {
	return &OrderHandler{uc: uc}
}

// RegisterRoutes registers all order routes on the provided router group.
func (h *OrderHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/orders", h.CreateOrder)
	rg.GET("/orders/:id", h.GetOrder)
	rg.PATCH("/orders/:id/cancel", h.CancelOrder)
	rg.GET("/orders", h.GetOrdersByCustomer)
}

type createOrderRequest struct {
	CustomerID string `json:"customer_id" binding:"required"`
	ItemName   string `json:"item_name"   binding:"required"`
	Amount     int64  `json:"amount"      binding:"required"`
}

type orderResponse struct {
	ID            string `json:"id"`
	CustomerID    string `json:"customer_id"`
	ItemName      string `json:"item_name"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
	TransactionID string `json:"transaction_id,omitempty"`
	CreatedAt     string `json:"created_at"`
}

// Handlers

// CreateOrder handles POST /orders
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	idempotencyKey := c.GetHeader("Idempotency-Key")

	output, err := h.uc.CreateOrder(usecase.CreateOrderInput{
		CustomerID:     req.CustomerID,
		ItemName:       req.ItemName,
		Amount:         req.Amount,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidAmount) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, domain.ErrPaymentServiceUnavailable) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "payment service unavailable"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	status := http.StatusCreated
	if !output.IsNew {
		status = http.StatusOK
	}

	c.JSON(status, orderResponse{
		ID:            output.Order.ID,
		CustomerID:    output.Order.CustomerID,
		ItemName:      output.Order.ItemName,
		Amount:        output.Order.Amount,
		Status:        output.Order.Status,
		TransactionID: output.TransactionID,
		CreatedAt:     output.Order.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// GetOrder handles GET /orders/:id
func (h *OrderHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")

	order, err := h.uc.GetOrder(id)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orderResponse{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		ItemName:   order.ItemName,
		Amount:     order.Amount,
		Status:     order.Status,
		CreatedAt:  order.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// CancelOrder handles PATCH /orders/:id/cancel
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	id := c.Param("id")

	order, err := h.uc.CancelOrder(id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOrderNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		case errors.Is(err, domain.ErrCancelPaidOrder):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, domain.ErrOrderNotCancellable):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, orderResponse{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		ItemName:   order.ItemName,
		Amount:     order.Amount,
		Status:     order.Status,
		CreatedAt:  order.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}
func (h *OrderHandler) GetOrdersByCustomer(c *gin.Context) {
	customerID := c.Query("customer_id")

	orders, err := h.uc.GetOrdersByCustomerID(customerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := make([]orderResponse, 0)

	for _, o := range orders {
		response = append(response, orderResponse{
			ID:         o.ID,
			CustomerID: o.CustomerID,
			ItemName:   o.ItemName,
			Amount:     o.Amount,
			Status:     o.Status,
			CreatedAt:  o.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, response)
}
