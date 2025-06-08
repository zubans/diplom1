package handlers

import (
	"bytes"
	"errors"
	"gophermart/internal/repos"
	"io"
	"net/http"

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	resp := make([]OrderResponse, 0, len(orders))
	for _, o := range orders {
		var accrual *float64
		if o.Accrual != 0 {
			accrual = &o.Accrual
		}
		resp = append(resp, OrderResponse{
			Number:     o.Number,
			Status:     o.Status,
			Accrual:    accrual,
			UploadedAt: o.UploadedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	c.JSON(http.StatusOK, resp)
}

func (h *OrderHandler) UploadOrder(c *gin.Context) {
	userID := c.GetInt("userID")

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	number := string(bytes.TrimSpace(body))

	status, err := h.orderService.AddOrder(c.Request.Context(), userID, number)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidNumber):
			c.AbortWithStatus(http.StatusUnprocessableEntity)
		case errors.Is(err, repos.ErrOrderConflict):
			c.AbortWithStatus(http.StatusConflict)
		default:
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		return
	}

	c.Status(status)
}
