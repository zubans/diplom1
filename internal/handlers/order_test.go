package handlers

import (
	"bytes"
	"context"
	"errors"
	"gophermart/internal/storage/repos"
	"gophermart/internal/storage/repos/mocks"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gophermart/internal/services"
)

func mockAuthMiddleware(userID int) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	}
}

func TestUploadOrder(t *testing.T) {
	mockRepo := new(mocks.OrderRepository)
	service := services.NewOrderService(mockRepo, "http://localhost")
	handler := NewOrderHandler(service)

	router := gin.Default()
	router.Use(mockAuthMiddleware(42))
	router.POST("/api/user/orders", handler.UploadOrder)

	tests := []struct {
		name       string
		body       string
		mockSetup  func(*mocks.OrderRepository)
		wantStatus int
	}{
		{
			name: "success new order",
			body: "12345678903",
			mockSetup: func(r *mocks.OrderRepository) {
				r.On("GetOrderUserID", mock.Anything, "12345678903").Return(0, nil).Once()
				r.On("CreateOrder", mock.Anything, 42, "12345678903").Return(nil).Once()
			},
			wantStatus: http.StatusAccepted,
		},
		{
			name: "order already uploaded by user",
			body: "12345678903",
			mockSetup: func(r *mocks.OrderRepository) {
				r.On("GetOrderUserID", mock.Anything, "12345678903").Return(42, nil).Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "order uploaded by another user",
			body: "12345678903",
			mockSetup: func(r *mocks.OrderRepository) {
				r.On("GetOrderUserID", mock.Anything, "12345678903").Return(99, nil).Once()
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "invalid order format",
			body: "invalid_order",
			mockSetup: func(r *mocks.OrderRepository) {
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "internal error",
			body: "12345678903",
			mockSetup: func(r *mocks.OrderRepository) {
				r.On("GetOrderUserID", mock.Anything, "12345678903").Return(0, nil).Once()
				r.On("CreateOrder", mock.Anything, 42, "12345678903").Return(errors.New("db error")).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(mockRepo)
			req := httptest.NewRequest("POST", "/api/user/orders", bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetOrders_InternalError(t *testing.T) {
	mockRepo := new(mocks.OrderRepository)
	service := services.NewOrderService(mockRepo, "http://localhost")
	handler := NewOrderHandler(service)

	router := gin.Default()
	router.Use(mockAuthMiddleware(1))
	router.GET("/orders", handler.GetOrders)

	mockRepo.On("GetOrders", mock.Anything, 1).Return([]repos.Order{}, errors.New("db error")).Once()

	req := httptest.NewRequest("GET", "/orders", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestGetOrders_NullAccrual(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repos.NewPostgresOrderRepository(db)
	userID := 1
	testNumber := "123"
	testStatus := "NEW"
	testUploadedAt := time.Now()

	rows := sqlmock.NewRows([]string{"number", "status", "accrual", "uploaded_at"}).
		AddRow(testNumber, testStatus, nil, testUploadedAt)

	mock.ExpectQuery("^SELECT (.+) FROM orders WHERE user_id = \\$1$").
		WithArgs(userID).
		WillReturnRows(rows)

	orders, err := repo.GetOrders(context.Background(), userID)

	assert.NoError(t, err)
	assert.Len(t, orders, 1)
	assert.Equal(t, testNumber, orders[0].Number)
	assert.Equal(t, testStatus, orders[0].Status)
	assert.Equal(t, 0.0, orders[0].Accrual)
	assert.WithinDuration(t, testUploadedAt, orders[0].UploadedAt, time.Millisecond)

	assert.NoError(t, mock.ExpectationsWereMet())
}
