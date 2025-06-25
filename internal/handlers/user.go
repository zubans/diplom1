package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gophermart/internal/dferrors"
	"gophermart/internal/dto"
	"gophermart/internal/services"
	"gophermart/pkg/logger"
	"net/http"
)

type AuthHandler struct {
	authService *services.AuthService
}

func New(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.CredentialsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Error("invalid request", zap.Error(err), zap.String("Login", req.Login), zap.Any("BODY", c.Request.Body))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	userID, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, dferrors.ErrUserAlreadyExists):
			logger.Log.Info("user already exists", zap.Error(err), zap.String("Login", req.Login), zap.Any("BODY", c.Request.Body))
			c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
		default:
			logger.Log.Info("internal error", zap.Error(err), zap.String("Login", req.Login), zap.Any("BODY", c.Request.Body))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	token, err := h.authService.GenerateToken(userID)
	if err != nil {
		logger.Log.Error("internal error", zap.Error(err), zap.String("Login", req.Login), zap.Any("BODY", c.Request.Body))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.Header("Authorization", "Bearer "+token)
	c.Status(http.StatusOK)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.CredentialsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Error("invalid request", zap.Error(err), zap.String("Login", req.Login), zap.Any("BODY", c.Request.Body))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	token, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, dferrors.ErrInvalidCredentials):
			logger.Log.Info("invalid credentials", zap.Error(err), zap.String("Login", req.Login), zap.Any("BODY", c.Request.Body))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		default:
			logger.Log.Error("internal error", zap.Error(err), zap.String("Login", req.Login), zap.Any("BODY", c.Request.Body))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.Header("Authorization", "Bearer "+token)
	c.Status(http.StatusOK)
}
