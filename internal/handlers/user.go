// handlers/auth_handler.go
package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gophermart/internal/dferrors"
	"gophermart/internal/services"
	"net/http"
)

type AuthHandler struct {
	authService *services.AuthService
}

func New(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	userID, err := h.authService.Register(c.Request.Context(), req.Login, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, dferrors.ErrUserAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	token, err := h.authService.GenerateToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.Header("Authorization", "Bearer "+token)
	c.Status(http.StatusOK)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	token, err := h.authService.Login(c.Request.Context(), req.Login, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, dferrors.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.Header("Authorization", "Bearer "+token)
	c.Status(http.StatusOK)
}
