package repos

import (
	"context"
	"database/sql"
	"time"
)

type WithdrawalRepository interface {
	GetWithdrawals(ctx context.Context, userID int) ([]Withdrawal, error)
}

type PostgresWithdrawalRepository struct {
	db *sql.DB
}

type Withdrawal struct {
	OrderNumber string
	Sum         float64
	ProcessedAt time.Time
}

func NewPostgresWithdrawalRepository(db *sql.DB) *PostgresWithdrawalRepository {
	return &PostgresWithdrawalRepository{db: db}
}

func (r *PostgresWithdrawalRepository) GetWithdrawals(ctx context.Context, userID int) ([]Withdrawal, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT order_number, sum, processed_at FROM withdrawals WHERE user_id = $1 ORDER BY processed_at`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Withdrawal
	for rows.Next() {
		var w Withdrawal
		if err := rows.Scan(&w.OrderNumber, &w.Sum, &w.ProcessedAt); err != nil {
			return nil, err
		}
		result = append(result, w)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
