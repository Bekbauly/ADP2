package domain

import (
	"errors"
	"time"
)

// Payment statuses
const (
	StatusAuthorized = "Authorized"
	StatusDeclined   = "Declined"
)

const MaxPaymentAmount int64 = 100000

type Payment struct {
	ID            string
	OrderID       string
	TransactionID string
	Amount        int64 // Amount in cents
	Status        string
	CreatedAt     time.Time
}

// Domain errors
var (
	ErrPaymentNotFound = errors.New("payment not found")
	ErrInvalidAmount   = errors.New("amount must be greater than 0")
)

func NewPayment(id, orderID, transactionID string, amount int64) (*Payment, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	status := StatusAuthorized
	if amount > MaxPaymentAmount {
		status = StatusDeclined
		transactionID = ""
	}

	return &Payment{
		ID:            id,
		OrderID:       orderID,
		TransactionID: transactionID,
		Amount:        amount,
		Status:        status,
		CreatedAt:     time.Now().UTC(),
	}, nil
}
