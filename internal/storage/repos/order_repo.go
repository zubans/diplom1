package repos

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Order struct {
	Number     string
	Status     string
	Accrual    float64
	UploadedAt time.Time
}

type OrderRepository interface {
	GetOrderUserID(ctx context.Context, number string) (int, error)
	CreateOrder(ctx context.Context, userID int, number string) error
	UpdateOrderStatus(ctx context.Context, number string, status string, accrual float64) error
	GetOrders(ctx context.Context, userID int) ([]Order, error)
	GetPendingOrders(ctx context.Context) ([]Order, error)
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

func (r *PostgresOrderRepository) GetOrders(ctx context.Context, userID int) ([]Order, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT number, status, accrual, uploaded_at FROM orders WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Order
	for rows.Next() {
		var o Order
		var accrual sql.NullFloat64
		if err := rows.Scan(&o.Number, &o.Status, &accrual, &o.UploadedAt); err != nil {
			return nil, err
		}

		if accrual.Valid {
			o.Accrual = accrual.Float64
		} else {
			o.Accrual = 0
		}

		result = append(result, o)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *PostgresOrderRepository) GetPendingOrders(ctx context.Context) ([]Order, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT number FROM orders 
		WHERE status NOT IN ('PROCESSED', 'INVALID')`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.Number); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	return orders, rows.Err()
}

//func (r *PostgresOrderRepository) GetOrders(ctx context.Context, userID int) ([]Order, error) {
//	rows, err := r.db.QueryContext(ctx,
//		`SELECT number, status, accrual, uploaded_at FROM orders WHERE user_id = $1`,
//		userID,
//	)
//	if err != nil {
//		return nil, err
//	}
//	defer rows.Close()
//
//	var result []Order
//	for rows.Next() {
//		var o Order
//		if err := rows.Scan(&o.Number, &o.Status, &o.Accrual, &o.UploadedAt); err != nil {
//			return nil, err
//		}
//		result = append(result, o)
//	}
//
//	if err := rows.Err(); err != nil {
//		return nil, err
//	}
//
//	return result, nil
//
//}

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

	if status == "PROCESSED" && accrual != 0 {
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
