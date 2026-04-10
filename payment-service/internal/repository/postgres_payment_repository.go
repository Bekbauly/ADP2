package repository

import (
	"database/sql"
	"fmt"
	"time"

	"payment-service/internal/domain"
)

// PostgresPaymentRepository is the implementation of domain.PaymentRepository.
type PostgresPaymentRepository struct {
	db *sql.DB
}

// NewPostgresPaymentRepository constructs the repository.
func NewPostgresPaymentRepository(db *sql.DB) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{db: db}
}

// Save inserts a new payment record.
func (r *PostgresPaymentRepository) Save(payment *domain.Payment) error {
	query := `
		INSERT INTO payments (id, order_id, transaction_id, amount, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(query,
		payment.ID,
		payment.OrderID,
		nullableString(payment.TransactionID),
		payment.Amount,
		payment.Status,
		payment.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("postgres save payment: %w", err)
	}
	return nil
}

// FindByOrderID retrieves a payment by its associated order ID.
func (r *PostgresPaymentRepository) FindByOrderID(orderID string) (*domain.Payment, error) {
	query := `
		SELECT id, order_id, COALESCE(transaction_id, ''), amount, status, created_at
		FROM payments
		WHERE order_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	row := r.db.QueryRow(query, orderID)

	var p domain.Payment
	var createdAt time.Time
	err := row.Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status, &createdAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrPaymentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("postgres find payment: %w", err)
	}
	p.CreatedAt = createdAt
	return &p, nil
}

// nullableString converts an empty string to NULL for optional fields.
func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
