package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"order-service/internal/domain"
)

// paymentRequestDTO is the wire format sent to Payment Service.
type paymentRequestDTO struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
}

// paymentResponseDTO is the wire format received from Payment Service.
type paymentResponseDTO struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
	Message       string `json:"message,omitempty"`
}

// PaymentHTTPClient is the concrete adapter that implements domain.PaymentClient.
type PaymentHTTPClient struct {
	httpClient     *http.Client
	paymentBaseURL string
}

// NewPaymentHTTPClient constructs the adapter.
func NewPaymentHTTPClient(client *http.Client, baseURL string) *PaymentHTTPClient {
	return &PaymentHTTPClient{
		httpClient:     client,
		paymentBaseURL: baseURL,
	}
}

// Authorize calls POST /payments on the Payment Service.
func (c *PaymentHTTPClient) Authorize(req domain.PaymentRequest) (*domain.PaymentResponse, error) {
	payload := paymentRequestDTO{
		OrderID: req.OrderID,
		Amount:  req.Amount,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payment request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.paymentBaseURL+"/payments",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		// Timeout or connection refused — triggers 503 in the handler
		return nil, fmt.Errorf("%w: %v", domain.ErrPaymentServiceUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("%w: payment service returned %d", domain.ErrPaymentServiceUnavailable, resp.StatusCode)
	}

	var dto paymentResponseDTO
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return nil, fmt.Errorf("decode payment response: %w", err)
	}

	return &domain.PaymentResponse{
		TransactionID: dto.TransactionID,
		Status:        dto.Status,
	}, nil
}
