-- Migration 001: Create payments table
-- Run this script against the "payments" database before starting the service.

CREATE TABLE IF NOT EXISTS payments (
    id             VARCHAR(36)  PRIMARY KEY,
    order_id       VARCHAR(36)  NOT NULL,
    transaction_id VARCHAR(36)  NULL,         -- NULL for declined payments
    amount         BIGINT       NOT NULL CHECK (amount > 0),  -- stored in cents
    status         VARCHAR(20)  NOT NULL,      -- "Authorized" or "Declined"
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);
CREATE INDEX IF NOT EXISTS idx_payments_status   ON payments(status);
