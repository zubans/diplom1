package handlers

import (
	"bytes"
	"errors"
	"go.uber.org/zap"
	"gophermart/internal/dferrors"
	"gophermart/pkg/logger"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gophermart/internal/services"
)

type OrderResponse struct {
	Number     string   `json:"number"`
	Status     string   `json:"status"`
	Accrual    *float64 `json:"accrual,omitempty"`
	UploadedAt string   `json:"uploaded_at"`
}

type OrderHandler struct {
	orderService *services.OrderService
}

func NewOrderHandler(orderService *services.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) GetOrders(c *gin.Context) {
	userID := c.GetInt("userID")

	orders, err := h.orderService.GetOrders(c.Request.Context(), userID)
	if err != nil {
		logger.Log.Error("internal error", zap.Error(err), zap.Int("UserID", userID), zap.Any("BODY", c.Request.Body))
		c.JSON(http.StatusOK, gin.H{"error": "internal error"})
		return
	}

	if len(orders) == 0 {
		c.Status(http.StatusOK)
		return
	}

	resp := make([]OrderResponse, 0, len(orders))
	for _, o := range orders {
		var accrual *float64
		if o.Accrual != 0 {
			accrual = &o.Accrual
		}

		uploadedAt := ""
		if !o.UploadedAt.IsZero() {
			uploadedAt = o.UploadedAt.Format(time.RFC3339)
		}
		resp = append(resp, OrderResponse{
			Number:     o.Number,
			Status:     o.Status,
			Accrual:    accrual,
			UploadedAt: uploadedAt,
		})
	}

	c.JSON(http.StatusOK, resp)
}

func (h *OrderHandler) UploadOrder(c *gin.Context) {
	userID := c.GetInt("userID")

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Log.Error("error read body", zap.Error(err), zap.Int("UserID", userID), zap.Any("BODY", c.Request.Body))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	number := string(bytes.TrimSpace(body))

	status, err := h.orderService.AddOrder(c.Request.Context(), userID, number)
	if err != nil {
		switch {
		case errors.Is(err, dferrors.ErrInvalidNumber):
			logger.Log.Error(dferrors.ErrInvalidNumber.Error(), zap.Error(err), zap.Int("UserID", userID), zap.Any("BODY", c.Request.Body))
			c.AbortWithStatus(http.StatusUnprocessableEntity)
		case errors.Is(err, dferrors.ErrOrderConflict):
			logger.Log.Error(dferrors.ErrOrderConflict.Error(), zap.Error(err), zap.Int("UserID", userID), zap.Any("BODY", c.Request.Body))
			c.AbortWithStatus(http.StatusConflict)
		default:
			logger.Log.Error("internal error", zap.Error(err), zap.Int("UserID", userID), zap.Any("BODY", c.Request.Body))
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		return
	}

	c.Status(status)
}
