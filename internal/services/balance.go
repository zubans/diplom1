package services

import (
	"context"
	"gophermart/internal/storage/repos"
)

type WithdrawRequest struct {
	UserID      int
	OrderNumber string
	Sum         float64
}

type BalanceService struct {
	repo        repos.BalanceRepository
	requestChan chan WithdrawRequest
	workerCount int
}

func NewBalanceService(balanceRepo repos.BalanceRepository, workers int) *BalanceService {
	s := &BalanceService{
		repo:        balanceRepo,
		requestChan: make(chan WithdrawRequest, 1000),
		workerCount: workers,
	}
	//s.startWorkers()
	return s
}

func (s *BalanceService) startWorkers() {
	for i := 0; i < s.workerCount; i++ {
		go s.ProcessWithdrawals()
	}
}

func (s *BalanceService) Withdraw(ctx context.Context, userID int, orderNumber string, sum float64) error {
	return s.repo.Withdraw(ctx, userID, orderNumber, sum)
}

func (s *BalanceService) ProcessWithdrawals() {
	//for req := range s.requestChan {
	//	ctx := context.Background()
	//	err := s.repo.Withdraw(ctx, req.UserID, req.OrderNumber, req.Sum)
	//	if err != nil {
	//		log.Printf("Withdrawal failed: %v", err)
	//	}
	//}
}

//func (s *BalanceService) processWithdrawals() {
//	for req := range s.requestChan {
//		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
//		err := s.repo.Withdraw(ctx, req.UserID, req.OrderNumber, req.Sum)
//		cancel()
//
//		if err != nil {
//			log.Printf("Withdrawal failed: %v", err)
//		}
//	}
//}

func (s *BalanceService) HasSufficientFunds(ctx context.Context, userID int, sum float64) (bool, error) {
	balance, err := s.repo.GetBalance(ctx, userID)
	if err != nil {
		return false, err
	}
	return balance.Current >= sum, nil
}

func (s *BalanceService) QueueWithdraw(req WithdrawRequest) {
	s.requestChan <- req
}

func (s *BalanceService) GetBalance(ctx context.Context, userID int) (*repos.Balance, error) {
	return s.repo.GetBalance(ctx, userID)
}
