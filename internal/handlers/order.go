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

type OrderHandler struct {
	orderService *services.OrderService
}

func NewOrderHandler(orderService *services.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
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
