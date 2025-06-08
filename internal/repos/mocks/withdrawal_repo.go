package mocks

import (
	"context"
	"gophermart/internal/repos"
)

func (m *BalanceRepository) Withdraw(ctx context.Context, userID int, orderNumber string, sum float64) error {
	args := m.Called(ctx, userID, orderNumber, sum)
	return args.Error(0)
}

func (m *BalanceRepository) GetWithdrawals(ctx context.Context, userID int) ([]repos.Withdrawal, error) {
	args := m.Called(ctx, userID)
	withdrawals, ok := args.Get(0).([]repos.Withdrawal)
	if !ok && args.Get(0) != nil {
		panic("GetWithdrawals: returned value is not []services.Withdrawal")
	}
	return withdrawals, args.Error(1)
}
