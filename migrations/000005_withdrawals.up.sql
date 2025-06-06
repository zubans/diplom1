CREATE TABLE withdrawals
(
    id           UUID PRIMARY KEY        DEFAULT gen_random_uuid(),
    user_id      INT            NOT NULL REFERENCES users (id),
    order_number VARCHAR(50)    NOT NULL UNIQUE,
    sum          NUMERIC(12, 2) NOT NULL CHECK (sum > 0),
    processed_at TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_loyalty_withdrawals_user ON withdrawals (user_id);