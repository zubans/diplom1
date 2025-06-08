package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ShiraazMoollatjie/goluhn"
	"gophermart/internal/repos"
)

var (
	ErrInvalidNumber = errors.New("invalid order number")
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
		return 0, ErrInvalidNumber
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
		return http.StatusConflict, repos.ErrOrderConflict
	}

	if err := s.orderRepo.CreateOrder(ctx, userID, number); err != nil {
		return 0, err
	}

	go s.startStatusChecker(number)

	return http.StatusAccepted, nil
}

func (s *OrderService) startStatusChecker(number string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			status, accrual, err := s.checkOrderStatus(ctx, number)
			if err != nil {
				log.Printf("Status check error: %v", err)
				continue
			}

			if err := s.orderRepo.UpdateOrderStatus(ctx, number, status, accrual); err != nil {
				log.Printf("Failed to update order status: %v", err)
			}

			if isFinalStatus(status) {
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

func (s *OrderService) GetOrders(ctx context.Context, userID int) ([]repos.Order, error) {
	orders, err := s.orderRepo.GetOrders(ctx, userID)

	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}
	return orders, nil

}

func (s *OrderService) checkOrderStatus(ctx context.Context, number string) (string, float64, error) {
	url := fmt.Sprintf("http://%s/api/orders/%s", s.accrualURL, number)

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

func validateOrderNumber(number string) error {
	if _, err := strconv.Atoi(number); err != nil {
		return ErrInvalidNumber
	}

	err := goluhn.Validate(number)
	if err != nil {
		return errors.New("invalid check digit")
	}
	return nil
}

func getCheckInterval() time.Duration {
	intervalStr := os.Getenv("CHECK_INTERVAL_SECONDS")
	if intervalStr == "" {
		intervalStr = "5"
	}

	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		return 5 * time.Second
	}
	return time.Duration(interval) * time.Second
}

func isFinalStatus(status string) bool {
	return status == "PROCESSED" || status == "INVALID"
}
