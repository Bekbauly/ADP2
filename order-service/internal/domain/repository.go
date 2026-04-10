package domain

// OrderRepository is the Port (interface) that use cases depend on.
type OrderRepository interface {
	Save(order *Order) error
	FindByID(id string) (*Order, error)
	Update(order *Order) error
	FindByIdempotencyKey(key string) (*Order, error)
	SaveWithIdempotencyKey(order *Order, key string) error
	FindByCustomerID(customerID string) ([]*Order, error)
}
