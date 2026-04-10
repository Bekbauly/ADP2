package usecase

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"order-service/internal/domain"
)

// OrderUseCase contains all business logic for orders.
type OrderUseCase struct {
	repo          domain.OrderRepository
	paymentClient domain.PaymentClient
}

// NewOrderUseCase is the constructor. Dependencies are injected from the Composition Root.
func NewOrderUseCase(repo domain.OrderRepository, paymentClient domain.PaymentClient) *OrderUseCase {
	return &OrderUseCase{
		repo:          repo,
		paymentClient: paymentClient,
	}
}

// CreateOrderInput is the DTO coming into the use case from the delivery layer.
type CreateOrderInput struct {
	CustomerID     string
	ItemName       string
	Amount         int64
	IdempotencyKey string // optional; for bonus idempotency
}

// CreateOrderOutput is the DTO returned to the delivery layer.
type CreateOrderOutput struct {
	Order         *domain.Order
	TransactionID string
	IsNew         bool
}

// CreateOrder orchestrates order creation and payment authorization.
func (uc *OrderUseCase) CreateOrder(input CreateOrderInput) (*CreateOrderOutput, error) {
	// --- Idempotency Check (Bonus) ---
	if input.IdempotencyKey != "" {
		existing, err := uc.repo.FindByIdempotencyKey(input.IdempotencyKey)
		if err == nil && existing != nil {
			return &CreateOrderOutput{Order: existing, IsNew: false}, nil
		}
		if err != nil && err != domain.ErrOrderNotFound {
			return nil, fmt.Errorf("check existing idempotency: %w", err)
		}
	}

	// --- Create domain entity (validation inside NewOrder) ---
	order, err := domain.NewOrder(uuid.NewString(), input.CustomerID, input.ItemName, input.Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid order: %w", err)
	}

	// --- Persist order with status "Pending" ---
	if input.IdempotencyKey != "" {
		if err := uc.repo.SaveWithIdempotencyKey(order, input.IdempotencyKey); err != nil {
			if errors.Is(err, domain.ErrDuplicateOrder) {
				existing, err2 := uc.repo.FindByIdempotencyKey(input.IdempotencyKey)
				if err2 != nil {
					return nil, fmt.Errorf("idempotent order lookup: %w", err2)
				}
				return &CreateOrderOutput{Order: existing, IsNew: false}, nil
			}
			return nil, fmt.Errorf("failed to save order: %w", err)
		}
	} else {
		if err := uc.repo.Save(order); err != nil {
			return nil, fmt.Errorf("failed to save order: %w", err)
		}
	}

	// --- Call Payment Service via Port ---
	payResp, err := uc.paymentClient.Authorize(domain.PaymentRequest{
		OrderID: order.ID,
		Amount:  order.Amount,
	})
	if err != nil {
		// Payment service unavailable — mark as Failed and persist
		order.MarkFailed()
		_ = uc.repo.Update(order)
		return nil, fmt.Errorf("%w: %v", domain.ErrPaymentServiceUnavailable, err)
	}

	// --- Apply payment result to order ---
	var transactionID string
	if payResp.Status == "Authorized" {
		order.MarkPaid()
		transactionID = payResp.TransactionID
	} else {
		order.MarkFailed()
	}

	if err := uc.repo.Update(order); err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	return &CreateOrderOutput{Order: order, TransactionID: transactionID, IsNew: true}, nil
}

// GetOrder retrieves an order by ID.
func (uc *OrderUseCase) GetOrder(id string) (*domain.Order, error) {
	order, err := uc.repo.FindByID(id)
	if err != nil {
		return nil, domain.ErrOrderNotFound
	}
	return order, nil
}

// CancelOrder cancels a pending order.
func (uc *OrderUseCase) CancelOrder(id string) (*domain.Order, error) {
	order, err := uc.repo.FindByID(id)
	if err != nil {
		return nil, domain.ErrOrderNotFound
	}

	if err := order.Cancel(); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(order); err != nil {
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	return order, nil
}

func (uc *OrderUseCase) GetOrdersByCustomerID(customerID string) ([]*domain.Order, error) {
	if customerID == "" {
		return nil, errors.New("customer_id is required")
	}

	return uc.repo.FindByCustomerID(customerID)
}
