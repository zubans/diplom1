package mocks

import (
	"context"
	"errors"
	"gophermart/internal/storage/repos"

	"github.com/stretchr/testify/mock"
)

type OrderRepository struct {
	mock.Mock
}

func (m *OrderRepository) GetOrders(ctx context.Context, userID int) ([]repos.Order, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]repos.Order), args.Error(1)
}

func (m *OrderRepository) GetOrderUserID(ctx context.Context, number string) (int, error) {
	args := m.Called(ctx, number)
	return args.Int(0), args.Error(1)
}

func (m *OrderRepository) CreateOrder(ctx context.Context, userID int, number string) error {
	args := m.Called(ctx, userID, number)
	return args.Error(0)
}

func (m *OrderRepository) UpdateOrderStatus(
	ctx context.Context,
	number string,
	status string,
	accrual float64,
) error {
	args := m.Called(ctx, number, status, accrual)
	return args.Error(0)
}

func (m *OrderRepository) MockSuccessCreate(number string, userID int) {
	m.On("CreateOrder", mock.Anything, userID, number).Return(nil)
}

func (m *OrderRepository) MockConflictCreate(number string, userID int) {
	m.On("CreateOrder", mock.Anything, userID, number).Return(errors.New("order conflict"))
}

func (m *OrderRepository) MockGetUserID(number string, userID int, err error) {
	m.On("GetOrderUserID", mock.Anything, number).Return(userID, err)
}

func (m *OrderRepository) MockUpdateStatus(number, status string, accrual float64, err error) {
	m.On("UpdateOrderStatus", mock.Anything, number, status, accrual).Return(err)
}
