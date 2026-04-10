package repository

import (
	"database/sql"
	"fmt"
	"time"

	"order-service/internal/domain"
)

// PostgresOrderRepository is the implementation of domain.
type PostgresOrderRepository struct {
	db *sql.DB
}

// NewPostgresOrderRepository constructs the repository with a DB connection.
func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

// Save inserts a new order into the database.
func (r *PostgresOrderRepository) Save(order *domain.Order) error {
	query := `
		INSERT INTO orders (id, customer_id, item_name, amount, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(query,
		order.ID,
		order.CustomerID,
		order.ItemName,
		order.Amount,
		order.Status,
		order.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("postgres save order: %w", err)
	}
	return nil
}

// FindByID retrieves an order by its primary key.
func (r *PostgresOrderRepository) FindByID(id string) (*domain.Order, error) {
	query := `
		SELECT id, customer_id, item_name, amount, status, created_at
		FROM orders
		WHERE id = $1
	`
	row := r.db.QueryRow(query, id)

	var o domain.Order
	var createdAt time.Time
	err := row.Scan(&o.ID, &o.CustomerID, &o.ItemName, &o.Amount, &o.Status, &createdAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("postgres find order: %w", err)
	}
	o.CreatedAt = createdAt
	return &o, nil
}

// Update persists changes to an existing order.
func (r *PostgresOrderRepository) Update(order *domain.Order) error {
	query := `
		UPDATE orders
		SET status = $1
		WHERE id = $2
	`
	result, err := r.db.Exec(query, order.Status, order.ID)
	if err != nil {
		return fmt.Errorf("postgres update order: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}

// FindByIdempotencyKey looks up an order by its idempotency key.
func (r *PostgresOrderRepository) FindByIdempotencyKey(key string) (*domain.Order, error) {
	query := `
		SELECT o.id, o.customer_id, o.item_name, o.amount, o.status, o.created_at
		FROM orders o
		INNER JOIN order_idempotency_keys ik ON ik.order_id = o.id
		WHERE ik.idempotency_key = $1
	`
	row := r.db.QueryRow(query, key)

	var o domain.Order
	var createdAt time.Time
	err := row.Scan(&o.ID, &o.CustomerID, &o.ItemName, &o.Amount, &o.Status, &createdAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("postgres find by idempotency key: %w", err)
	}
	o.CreatedAt = createdAt
	return &o, nil
}

// SaveWithIdempotencyKey saves an order and its idempotency key atomically.
func (r *PostgresOrderRepository) SaveWithIdempotencyKey(order *domain.Order, key string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Guard for duplicate idempotency key
	var existingOrderID string
	if err := tx.QueryRow(`
		SELECT order_id FROM order_idempotency_keys WHERE idempotency_key = $1
	`, key).Scan(&existingOrderID); err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("check idempotency key: %w", err)
		}
	} else {
		return domain.ErrDuplicateOrder
	}

	_, err = tx.Exec(`
		INSERT INTO orders (id, customer_id, item_name, amount, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, order.ID, order.CustomerID, order.ItemName, order.Amount, order.Status, order.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert order in tx: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO order_idempotency_keys (idempotency_key, order_id)
		VALUES ($1, $2)
	`, key, order.ID)
	if err != nil {
		return fmt.Errorf("insert idempotency key: %w", err)
	}

	return tx.Commit()
}

func (r *PostgresOrderRepository) FindByCustomerID(customerID string) ([]*domain.Order, error) {
	query := `
		SELECT id, customer_id, item_name, amount, status, created_at
		FROM orders
		WHERE customer_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, customerID)
	if err != nil {
		return nil, fmt.Errorf("postgres find orders by customer: %w", err)
	}
	defer rows.Close()

	var orders []*domain.Order

	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(
			&o.ID,
			&o.CustomerID,
			&o.ItemName,
			&o.Amount,
			&o.Status,
			&o.CreatedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, &o)
	}

	return orders, nil
}
