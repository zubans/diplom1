package mocks

import (
	"context"
	"github.com/stretchr/testify/mock"
	"gophermart/internal/storage/repos"
)

type BalanceRepository struct {
	mock.Mock
}

func (m *BalanceRepository) GetBalance(ctx context.Context, userID int) (*repos.Balance, error) {
	args := m.Called(ctx, userID)
	balance, ok := args.Get(0).(*repos.Balance)
	if !ok && args.Get(0) != nil {
		panic("GetBalance: returned value is not *services.Balance")
	}
	return balance, args.Error(1)
}
