package repos

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrOrderExists   = errors.New("order exists")
	ErrOrderConflict = errors.New("order conflict")
)

type OrderRepository interface {
	GetOrderUserID(ctx context.Context, number string) (int, error)
	CreateOrder(ctx context.Context, userID int, number string) error
	UpdateOrderStatus(ctx context.Context, number string, status string, accrual float64) error
}

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) GetOrderUserID(ctx context.Context, number string) (int, error) {
	var userID int
	query := `SELECT user_id FROM orders WHERE number = $1`
	err := r.db.QueryRowContext(ctx, query, number).Scan(&userID)

	if err == sql.ErrNoRows {
		return 0, nil
	}
	return userID, err
}

func (r *PostgresOrderRepository) CreateOrder(ctx context.Context, userID int, number string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO orders (number, user_id, status) 
		VALUES ($1, $2, 'NEW') 
		ON CONFLICT (number) DO NOTHING`,
		number, userID)

	if err != nil {
		return err
	}
	return nil
}

func (r *PostgresOrderRepository) UpdateOrderStatus(ctx context.Context, number string, status string, accrual float64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`UPDATE orders 
         SET status = $1, 
             accrual = $2 
         WHERE number = $3`,
		status,
		accrual,
		number,
	)

	if err != nil {
		return err
	}

	if "PROCESSED" == status {
		var userID int
		err = tx.QueryRowContext(ctx, `SELECT user_id FROM orders WHERE number = $1 FOR UPDATE`, number).Scan(&userID)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx,
			`UPDATE wallets 
             SET current_balance = current_balance + $1 
             WHERE user_id = $2`,
			accrual,
			userID,
		)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx,
			`INSERT INTO transactions 
             (user_id, type, amount, order_id, status) 
             VALUES ($1, 'ACCRUAL', $2, 
             (SELECT id FROM orders WHERE number = $3), 'COMPLETED')`,
			userID,
			accrual,
			number,
		)
		if err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return err
}
