package services

import (
	"context"
	"gophermart/internal/repos"
)

type WithdrawalService struct {
	repo repos.WithdrawalRepository
}

func NewWithdrawalService(withdrawalRepo repos.WithdrawalRepository) *WithdrawalService {
	return &WithdrawalService{
		repo: withdrawalRepo,
	}
}

func (s *WithdrawalService) GetWithdrawals(ctx context.Context, userID int) ([]repos.Withdrawal, error) {
	return s.repo.GetWithdrawals(ctx, userID)
}
