package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/mock"
	"gophermart/internal/dferrors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gophermart/internal/repos/mocks"
	"gophermart/internal/services"
)

func TestRegisterHandler(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	authService := services.NewAuthService(userRepo, "secret")
	handler := New(authService)

	r := gin.Default()
	r.POST("/register", handler.Register)

	tests := []struct {
		name      string
		payload   interface{}
		mockSetup func(*mocks.UserRepository)
		wantCode  int
	}{
		{
			name: "successful registration",
			payload: map[string]string{
				"login":    "testuser",
				"password": "validpass",
			},
			mockSetup: func(m *mocks.UserRepository) {
				m.On("CreateUser", mock.Anything, "testuser", mock.Anything).
					Return(1, nil).Once()
			},
			wantCode: http.StatusOK,
		},
		{
			name: "duplicate user",
			payload: map[string]string{
				"login":    "existinguser",
				"password": "validpass",
			},
			mockSetup: func(m *mocks.UserRepository) {
				m.On("CreateUser", mock.Anything, "existinguser", mock.Anything).
					Return(0, dferrors.ErrUserAlreadyExists).Once()
			},
			wantCode: http.StatusConflict,
		},
		{
			name:      "invalid payload",
			payload:   "invalid",
			mockSetup: func(m *mocks.UserRepository) {},
			wantCode:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(userRepo)
			body, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantCode, w.Code)
			userRepo.AssertExpectations(t)
		})
	}
}
