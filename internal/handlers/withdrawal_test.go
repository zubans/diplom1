package handlers

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gophermart/internal/services"
	"gophermart/internal/storage/repos"
	"gophermart/internal/storage/repos/mocks"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetWithdrawals(t *testing.T) {
	mockRepo := new(mocks.BalanceRepository)
	withdrawalService := services.NewWithdrawalService(mockRepo)
	withdrawalHandler := NewWithdrawalHandler(withdrawalService)

	router := gin.Default()
	router.Use(mockAuthMiddleware(77))
	router.GET("/api/user/withdrawals", withdrawalHandler.GetWithdrawals)

	now := time.Now()

	tests := []struct {
		name         string
		mockSetup    func(*mocks.BalanceRepository)
		wantStatus   int
		wantResponse []WithdrawalResponse
	}{
		{
			name: "success",
			mockSetup: func(r *mocks.BalanceRepository) {
				r.On("GetWithdrawals", mock.Anything, 77).
					Return([]repos.Withdrawal{
						{
							OrderNumber: "2377225624",
							Sum:         751,
							ProcessedAt: now,
						},
						{
							OrderNumber: "2377225625",
							Sum:         100,
							ProcessedAt: now.Add(2 * time.Hour),
						},
					}, nil).Once()
			},
			wantStatus: http.StatusOK,
			wantResponse: []WithdrawalResponse{
				{
					Order:       "2377225624",
					Sum:         751,
					ProcessedAt: now.Format("2006-01-02T15:04:05Z07:00"),
				},
				{
					Order:       "2377225625",
					Sum:         100,
					ProcessedAt: now.Add(2 * time.Hour).Format("2006-01-02T15:04:05Z07:00"),
				},
			},
		},
		{
			name: "internal error",
			mockSetup: func(r *mocks.BalanceRepository) {
				r.On("GetWithdrawals", mock.Anything, 77).
					Return(nil, errors.New("db error")).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "empty list",
			mockSetup: func(r *mocks.BalanceRepository) {
				r.On("GetWithdrawals", mock.Anything, 77).
					Return([]repos.Withdrawal{}, nil).Once()
			},
			wantStatus:   http.StatusOK,
			wantResponse: []WithdrawalResponse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(mockRepo)
			req := httptest.NewRequest("GET", "/api/user/withdrawals", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				var resp []WithdrawalResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResponse, resp)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
