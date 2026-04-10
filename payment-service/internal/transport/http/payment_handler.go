package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"payment-service/internal/domain"
	"payment-service/internal/usecase"
)

type PaymentHandler struct {
	uc *usecase.PaymentUseCase
}

// NewPaymentHandler constructs the handler.
func NewPaymentHandler(uc *usecase.PaymentUseCase) *PaymentHandler {
	return &PaymentHandler{uc: uc}
}

// RegisterRoutes wires all payment routes.
func (h *PaymentHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/payments", h.Authorize)
	rg.GET("/payments/:order_id", h.GetByOrderID)
}

type authorizeRequest struct {
	OrderID string `json:"order_id" binding:"required"`
	Amount  int64  `json:"amount"   binding:"required"`
}

type paymentResponse struct {
	ID            string `json:"id"`
	OrderID       string `json:"order_id"`
	TransactionID string `json:"transaction_id,omitempty"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

// Authorize handles POST /payments
func (h *PaymentHandler) Authorize(c *gin.Context) {
	var req authorizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	output, err := h.uc.Authorize(usecase.AuthorizeInput{
		OrderID: req.OrderID,
		Amount:  req.Amount,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidAmount) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	statusCode := http.StatusCreated
	if output.Payment.Status == domain.StatusDeclined {
		statusCode = http.StatusPaymentRequired // 402 for declined
	}

	c.JSON(statusCode, paymentResponse{
		ID:            output.Payment.ID,
		OrderID:       output.Payment.OrderID,
		TransactionID: output.Payment.TransactionID,
		Amount:        output.Payment.Amount,
		Status:        output.Payment.Status,
		CreatedAt:     output.Payment.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// GetByOrderID handles GET /payments/:order_id
func (h *PaymentHandler) GetByOrderID(c *gin.Context) {
	orderID := c.Param("order_id")

	payment, err := h.uc.GetByOrderID(orderID)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, paymentResponse{
		ID:            payment.ID,
		OrderID:       payment.OrderID,
		TransactionID: payment.TransactionID,
		Amount:        payment.Amount,
		Status:        payment.Status,
		CreatedAt:     payment.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}
