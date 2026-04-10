package usecase

import (
	"fmt"

	"github.com/google/uuid"

	"payment-service/internal/domain"
)

// PaymentUseCase contains all business logic for payments.
type PaymentUseCase struct {
	repo domain.PaymentRepository
}

// NewPaymentUseCase is the constructor with injected repository port.
func NewPaymentUseCase(repo domain.PaymentRepository) *PaymentUseCase {
	return &PaymentUseCase{repo: repo}
}

// AuthorizeInput is the DTO coming into the use case.
type AuthorizeInput struct {
	OrderID string
	Amount  int64
}

// AuthorizeOutput is the DTO returned to the delivery layer.
type AuthorizeOutput struct {
	Payment *domain.Payment
}

// Authorize processes a payment authorization request.
// Business rules:
//   - Amount must be > 0 (enforced in domain.NewPayment)
//   - Amount > 100000 cents → Declined (enforced in domain.NewPayment)
func (uc *PaymentUseCase) Authorize(input AuthorizeInput) (*AuthorizeOutput, error) {
	transactionID := uuid.NewString()
	paymentID := uuid.NewString()

	payment, err := domain.NewPayment(paymentID, input.OrderID, transactionID, input.Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid payment: %w", err)
	}

	if err := uc.repo.Save(payment); err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	return &AuthorizeOutput{Payment: payment}, nil
}

// GetByOrderID retrieves a payment by order ID.
func (uc *PaymentUseCase) GetByOrderID(orderID string) (*domain.Payment, error) {
	payment, err := uc.repo.FindByOrderID(orderID)
	if err != nil {
		return nil, domain.ErrPaymentNotFound
	}
	return payment, nil
}
