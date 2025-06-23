package repos

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"go.uber.org/zap"
	"gophermart/internal/dferrors"
	"gophermart/pkg/logger"
)

type PostgresBalanceRepository struct {
	db *sql.DB
}

type Balance struct {
	Current   float64
	Withdrawn float64
}

type BalanceRepository interface {
	GetBalance(ctx context.Context, userID int) (*Balance, error)
	Withdraw(ctx context.Context, userID int, orderNumber string, sum float64) error
}

func NewPostgresBalanceRepository(db *sql.DB) *PostgresBalanceRepository {
	return &PostgresBalanceRepository{db: db}
}

func (r *PostgresBalanceRepository) GetBalance(ctx context.Context, userID int) (*Balance, error) {
	var b Balance
	err := r.db.QueryRowContext(ctx,
		`SELECT current_balance, withdrawn_balance FROM wallets WHERE user_id = $1`,
		userID,
	).Scan(&b.Current, &b.Withdrawn)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *PostgresBalanceRepository) Withdraw(ctx context.Context, userID int, orderNumber string, sum float64) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})

	if err != nil {
		return fmt.Errorf("transaction begin error: %w", err)
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
			logger.Log.Error("Error rollback transaction withdraw", zap.Error(err), zap.Int("UserID", userID), zap.String("orderNumber", orderNumber))
		}
	}(tx)

	var currentBalance float64
	err = tx.QueryRowContext(ctx,
		`SELECT current_balance 
				FROM wallets 
				WHERE user_id = $1 FOR UPDATE`, userID).Scan(&currentBalance)

	if err != nil {
		return fmt.Errorf("balance lock failed: %w", err)
	}

	if currentBalance < sum {
		return dferrors.ErrInsufficientFunds
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO withdrawals (user_id, order_number, sum) 
				VALUES ($1, $2, $3)`, userID, orderNumber, sum)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return dferrors.ErrDuplicateWithdrawal
		}
		return fmt.Errorf("withdrawal insert error: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE wallets 
				SET current_balance = current_balance - $1, withdrawn_balance = withdrawn_balance + $1 
				WHERE user_id = $2`, sum, userID)

	if err != nil {
		return fmt.Errorf("balance update error: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO transactions (user_id, type, amount, order_id, status)
				VALUES ($1, $2, $3, $4, $5)`, userID, "WITHDRAWAL", sum, orderNumber, "COMPLETED")

	if err != nil {
		return fmt.Errorf("balance update error: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("transaction commit error: %w", err)
	}

	return nil
}
