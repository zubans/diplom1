package handlers

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gophermart/internal/repos"
	"gophermart/internal/repos/mocks"
	"gophermart/internal/services"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetBalance(t *testing.T) {
	mockRepo := new(mocks.BalanceRepository)
	balanceService := services.NewBalanceService(mockRepo, 10)
	balanceHandler := NewBalanceHandler(balanceService)

	router := gin.Default()
	router.Use(mockAuthMiddleware(77))
	router.GET("/api/user/balance", balanceHandler.GetBalance)

	tests := []struct {
		name          string
		mockSetup     func(*mocks.BalanceRepository)
		wantStatus    int
		wantCurrent   float64
		wantWithdrawn float64
	}{
		{
			name: "success",
			mockSetup: func(r *mocks.BalanceRepository) {
				r.On("GetBalance", mock.Anything, 77).
					Return(&repos.Balance{Current: 500.5, Withdrawn: 200.0}, nil).Once()
			},
			wantStatus:    http.StatusOK,
			wantCurrent:   500.5,
			wantWithdrawn: 200.0,
		},
		{
			name: "internal error",
			mockSetup: func(r *mocks.BalanceRepository) {
				r.On("GetBalance", mock.Anything, 77).
					Return(nil, errors.New("db error")).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(mockRepo)
			req := httptest.NewRequest("GET", "/api/user/balance", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				var resp struct {
					Current   float64 `json:"current"`
					Withdrawn float64 `json:"withdrawn"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCurrent, resp.Current)
				assert.Equal(t, tt.wantWithdrawn, resp.Withdrawn)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
