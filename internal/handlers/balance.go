package handlers

import (
	"github.com/gin-gonic/gin"
	"gophermart/internal/services"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	h.balanceService.QueueWithdraw(services.WithdrawRequest{
		UserID:      c.GetInt("userID"),
		OrderNumber: req.Order,
		Sum:         req.Sum,
	})

	c.Status(http.StatusAccepted)
}
