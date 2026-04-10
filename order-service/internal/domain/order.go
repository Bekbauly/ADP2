package domain

import (
	"errors"
	"time"
)

// Order status
const (
	StatusPending   = "Pending"
	StatusPaid      = "Paid"
	StatusFailed    = "Failed"
	StatusCancelled = "Cancelled"
)

type Order struct {
	ID         string
	CustomerID string
	ItemName   string
	Amount     int64
	Status     string
	CreatedAt  time.Time
}

// Domain error
var (
	ErrOrderNotFound       = errors.New("order not found")
	ErrInvalidAmount       = errors.New("amount must be greater than 0")
	ErrCancelPaidOrder     = errors.New("paid orders cannot be cancelled")
	ErrOrderNotCancellable = errors.New("only pending orders can be cancelled")
	ErrDuplicateOrder      = errors.New("duplicate order: idempotency key already used")
)

// New Order creates and validates a new Order domain.
func NewOrder(id, customerID, itemName string, amount int64) (*Order, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}
	return &Order{
		ID:         id,
		CustomerID: customerID,
		ItemName:   itemName,
		Amount:     amount,
		Status:     StatusPending,
		CreatedAt:  time.Now().UTC(),
	}, nil
}

// Cancel function
func (o *Order) Cancel() error {
	if o.Status == StatusPaid {
		return ErrCancelPaidOrder
	}
	if o.Status != StatusPending {
		return ErrOrderNotCancellable
	}
	o.Status = StatusCancelled
	return nil
}

// MarkPaid transitions an order to Paid status.
func (o *Order) MarkPaid() {
	o.Status = StatusPaid
}

// MarkFailed transitions an order to Failed status.
func (o *Order) MarkFailed() {
	o.Status = StatusFailed
}
