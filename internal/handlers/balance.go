package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gophermart/internal/dferrors"
	"gophermart/internal/services"
	"gophermart/pkg/logger"
	"net/http"
)

type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type BalanceHandler struct {
	balanceService *services.BalanceService
}

func NewBalanceHandler(balanceService *services.BalanceService) *BalanceHandler {
	return &BalanceHandler{balanceService: balanceService}
}

func (h *BalanceHandler) GetBalance(c *gin.Context) {
	userID := c.GetInt("userID")

	balance, err := h.balanceService.GetBalance(c.Request.Context(), userID)
	if err != nil {
		logger.Log.Error("Failed to get balance", zap.Error(err), zap.Int("userID", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, BalanceResponse{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	})
}

func (h *BalanceHandler) Withdraw(c *gin.Context) {
	var req struct {
		Order string  `json:"order" binding:"required"`
		Sum   float64 `json:"sum" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Error("Failed to withdraw", zap.Error(err), zap.String("Order", req.Order))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	userID := c.GetInt("userID")

	hasFunds, err := h.balanceService.HasSufficientFunds(c.Request.Context(), userID, req.Sum)
	if err != nil {
		logger.Log.Error("Failed to withdraw", zap.Error(err), zap.String("Order", req.Order), zap.Int("UserID", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if !hasFunds {
		logger.Log.Error("insufficient funds", zap.Error(err), zap.Int("UserID", userID))
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient funds"})
		return
	}

	err = h.balanceService.Withdraw(c.Request.Context(), userID, req.Order, req.Sum)
	switch {
	case errors.Is(err, dferrors.ErrInsufficientFunds):
		logger.Log.Error("insufficient funds", zap.Error(err), zap.Int("UserID", userID))
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient funds"})
	case errors.Is(err, dferrors.ErrDuplicateWithdrawal):
		logger.Log.Error("duplicate order", zap.Error(err), zap.String("OrderID", req.Order))
		c.JSON(http.StatusConflict, gin.H{"error": "duplicate order"})
	case err != nil:
		logger.Log.Error("duplicate order", zap.Error(err), zap.String("OrderID", req.Order), zap.Int("UserID", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	default:
		c.Status(http.StatusOK)
	}

	//h.balanceService.QueueWithdraw(services.WithdrawRequest{
	//	UserID:      c.GetInt("userID"),
	//	OrderNumber: req.Order,
	//	Sum:         req.Sum,
	//})
	//
	//c.Status(http.StatusOK)
}
