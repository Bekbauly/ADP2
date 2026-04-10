package domain

// PaymentRepository is the Port that use cases depend on.
type PaymentRepository interface {
	Save(payment *Payment) error
	FindByOrderID(orderID string) (*Payment, error)
}
