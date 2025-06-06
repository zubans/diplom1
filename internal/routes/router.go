package routes

import (
	"github.com/gin-gonic/gin"
	"gophermart/internal/handlers"
)

func SetupUserRoutes(r *gin.Engine, userHandler *handlers.AuthHandler) {
	api := r.Group("/api/user")
	{
		api.POST("/register", userHandler.Register)
		api.POST("/login", userHandler.Login)
	}
}

func SetupOrderRoutes(r *gin.Engine, orderHandler *handlers.OrderHandler, authMW gin.HandlerFunc) {
	orderGroup := r.Group("/api/user/orders")
	orderGroup.Use(authMW)
	{
		orderGroup.POST("", orderHandler.UploadOrder)
	}
}
