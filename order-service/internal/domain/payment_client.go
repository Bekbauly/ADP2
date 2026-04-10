package domain

import "errors"

// PaymentRequest is the outbound DTO sent to Payment Service.
type PaymentRequest struct {
	OrderID string
	Amount  int64
}

// PaymentRequest is the outgoing DTO
type PaymentResponse struct {
	TransactionID string
	Status        string //
}

// PaymentClient is the Port for outbound HTTP communication.
type PaymentClient interface {
	Authorize(req PaymentRequest) (*PaymentResponse, error)
}

// Payment client errors
var (
	ErrPaymentServiceUnavailable = errors.New("payment service unavailable")
)
