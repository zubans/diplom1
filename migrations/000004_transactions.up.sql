CREATE TABLE transactions
(
    id         UUID PRIMARY KEY        DEFAULT gen_random_uuid(),
    user_id    INT            NOT NULL REFERENCES users (id),
    type       VARCHAR(20)    NOT NULL CHECK (type IN ('ACCRUAL', 'WITHDRAWAL')),
    amount     NUMERIC(12, 2) NOT NULL CHECK (amount >= 0),
    order_id   BIGINT NOT NULL,
    status     VARCHAR(20)    NOT NULL CHECK (status IN ('PENDING', 'COMPLETED', 'CANCELED')),
    created_at TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_loyalty_transactions_user ON transactions (user_id);
CREATE INDEX idx_loyalty_transactions_order ON transactions (order_id);