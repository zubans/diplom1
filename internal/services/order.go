package services

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"gophermart/internal/dferrors"
	"gophermart/internal/storage/repos"
	"gophermart/pkg/logger"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ShiraazMoollatjie/goluhn"
)

type OrderService struct {
	orderRepo     repos.OrderRepository
	accrualURL    string
	client        *http.Client
	checkInterval time.Duration
}

func NewOrderService(orderRepo repos.OrderRepository, accrualURL string) *OrderService {
	return &OrderService{
		orderRepo:     orderRepo,
		accrualURL:    accrualURL,
		client:        &http.Client{Timeout: 5 * time.Second},
		checkInterval: getCheckInterval(),
	}
}

func (s *OrderService) AddOrder(ctx context.Context, userID int, number string) (int, error) {
	if _, err := strconv.Atoi(number); err != nil {
		return 0, dferrors.ErrInvalidNumber
	}

	if err := validateOrderNumber(number); err != nil {
		return http.StatusBadRequest, err
	}

	existingUserID, err := s.orderRepo.GetOrderUserID(ctx, number)
	if err != nil {
		return 0, err
	}

	switch {
	case existingUserID == userID:
		return http.StatusOK, nil
	case existingUserID != 0:
		return http.StatusConflict, dferrors.ErrOrderConflict
	}

	if err := s.orderRepo.CreateOrder(ctx, userID, number); err != nil {
		return 0, err
	}

	status, accrual, err := s.checkOrderStatus(ctx, number)
	if err == nil {
		if isFinalStatus(status) {
			if err := s.orderRepo.UpdateOrderStatus(ctx, number, status, accrual); err != nil {
				return 0, err
			}
			return mapStatusToHTTP(status), nil
		}
	}

	go func() {
		if err := s.startStatusChecker(ctx, number); err != nil {
			fmt.Printf("Status checker error: %v\n", err)
		}
	}()

	return http.StatusAccepted, nil
}

func (s *OrderService) startStatusChecker(ctx context.Context, number string) error {
	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	var (
		retries204    int
		maxRetries204 = 3
	)

	for {
		select {
		case <-ticker.C:
			status, accrual, err := s.checkOrderStatus(ctx, number)
			if err != nil {
				return fmt.Errorf("status check error: %w", err)
			}

			if status == "204" {
				retries204++
				if retries204 >= maxRetries204 {
					if err := s.orderRepo.UpdateOrderStatus(ctx, number, "INVALID", 0); err != nil {
						return fmt.Errorf("mark order invalid error: %w", err)
					}
					return nil
				}
				continue
			}
			retries204 = 0

			if err := s.orderRepo.UpdateOrderStatus(ctx, number, status, accrual); err != nil {
				return fmt.Errorf("update status error: %w", err)
			}

			if isFinalStatus(status) {
				return nil
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *OrderService) GetOrders(ctx context.Context, userID int) ([]repos.Order, error) {
	orders, err := s.orderRepo.GetOrders(ctx, userID)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (s *OrderService) checkOrderStatus(ctx context.Context, number string) (string, float64, error) {
	url := fmt.Sprintf("%s/api/orders/%s", s.accrualURL, number)

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := s.client.Do(req)
	if err != nil {
		return "", 0, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Status  string  `json:"status"`
		Accrual float64 `json:"accrual"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, err
	}

	return result.Status, result.Accrual, nil
}

func (s *OrderService) RecoverPendingOrders(ctx context.Context) error {
	orders, err := s.orderRepo.GetPendingOrders(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pending orders: %w", err)
	}

	for _, order := range orders {
		s.QueueOrderCheck(ctx, order.Number)
	}

	return nil
}

func (s *OrderService) QueueOrderCheck(ctx context.Context, number string) {
	go func() {
		err := s.startStatusChecker(ctx, number)
		if err != nil {
			logger.Log.Error("Error starting status checker", zap.Error(err))
		}
	}()
}

func validateOrderNumber(number string) error {
	if _, err := strconv.Atoi(number); err != nil {
		return dferrors.ErrInvalidNumber
	}

	err := goluhn.Validate(number)
	if err != nil {
		return dferrors.ErrInvalidNumber
	}
	return nil
}

func getCheckInterval() time.Duration {
	intervalStr := os.Getenv("CHECK_INTERVAL_SECONDS")
	if intervalStr == "" {
		intervalStr = "1"
	}

	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		return 1 * time.Second
	}
	return time.Duration(interval) * time.Second
}

func isFinalStatus(status string) bool {
	return status == "PROCESSED" || status == "INVALID"
}

func mapStatusToHTTP(status string) int {
	switch status {
	case "PROCESSED", "INVALID":
		return http.StatusAccepted
	default:
		return http.StatusOK
	}
}
