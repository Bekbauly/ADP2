
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS orders (
    id          VARCHAR(36)  PRIMARY KEY,
    customer_id VARCHAR(36)  NOT NULL,
    item_name   VARCHAR(255) NOT NULL,
    amount      BIGINT       NOT NULL CHECK (amount > 0),
    status      VARCHAR(20)  NOT NULL DEFAULT 'Pending',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_id);
CREATE INDEX IF NOT EXISTS idx_orders_status      ON orders(status);

-- Idempotency keys table (Bonus feature)
CREATE TABLE IF NOT EXISTS order_idempotency_keys (
    idempotency_key VARCHAR(255) PRIMARY KEY,
    order_id        VARCHAR(36)  NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
