CREATE TABLE orders (
                        id            SERIAL PRIMARY KEY,
                        number        VARCHAR(32) NOT NULL UNIQUE,
                        user_id       INTEGER NOT NULL REFERENCES users(id),
                        status        VARCHAR(16) NOT NULL,         -- NEW, PROCESSING, INVALID, PROCESSED
                        accrual       NUMERIC(12,2),
                        uploaded_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                        CONSTRAINT unique_user_number UNIQUE (user_id, number)
);
