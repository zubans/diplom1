package mocks

import (
	"context"
	"gophermart/internal/repos"

	"github.com/stretchr/testify/mock"
)

type UserRepository struct {
	mock.Mock
}

func (m *UserRepository) CreateUser(ctx context.Context, login, passwordHash string) (int, error) {
	args := m.Called(ctx, login, passwordHash)
	return args.Int(0), args.Error(1)
}

func (m *UserRepository) GetPasswordHash(ctx context.Context, login string) (string, error) {
	args := m.Called(ctx, login)
	return args.String(0), args.Error(1)
}

func (m *UserRepository) GetUserByLogin(ctx context.Context, login string) (*repos.User, error) {
	args := m.Called(ctx, login)
	return args.Get(0).(*repos.User), args.Error(1)
}

func (m *UserRepository) UpdateBalance(ctx context.Context, userID int, amount float64) error {
	args := m.Called(ctx, userID, amount)
	return args.Error(0)
}
