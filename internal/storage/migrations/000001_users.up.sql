CREATE TABLE users (
                       id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
                       username VARCHAR(50) UNIQUE NOT NULL,
                       password_hash VARCHAR(60) NOT NULL,
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
