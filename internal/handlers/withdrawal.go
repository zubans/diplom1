package handlers

import (
	"github.com/gin-gonic/gin"
	"gophermart/internal/services"
	"net/http"
)

type WithdrawalResponse struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

type WithdrawalHandler struct {
	WithdrawalService *services.WithdrawalService
}

func NewWithdrawalHandler(withdrawalService *services.WithdrawalService) *WithdrawalHandler {
	return &WithdrawalHandler{WithdrawalService: withdrawalService}
}

func (h *WithdrawalHandler) GetWithdrawals(c *gin.Context) {
	userID := c.GetInt("userID")

	withdrawals, err := h.WithdrawalService.GetWithdrawals(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	resp := make([]WithdrawalResponse, 0, len(withdrawals))
	for _, w := range withdrawals {
		resp = append(resp, WithdrawalResponse{
			Order:       w.OrderNumber,
			Sum:         w.Sum,
			ProcessedAt: w.ProcessedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	c.JSON(http.StatusOK, resp)
}
