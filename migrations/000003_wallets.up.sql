CREATE TABLE wallets
(
    id                SERIAL PRIMARY KEY,
    user_id           INT            NOT NULL UNIQUE REFERENCES users (id),
    current_balance   NUMERIC(12, 2) NOT NULL DEFAULT 0 CHECK (current_balance >= 0),
    withdrawn_balance NUMERIC(12, 2) NOT NULL DEFAULT 0 CHECK (withdrawn_balance >= 0),
    created_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);