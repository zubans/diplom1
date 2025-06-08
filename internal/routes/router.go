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
		orderGroup.GET("", orderHandler.GetOrders)
	}
}

func SetupBalanceRoutes(r *gin.Engine, balanceHandler *handlers.BalanceHandler, authMW gin.HandlerFunc) {
	api := r.Group("/api/user")
	api.Use(authMW)
	{
		api.GET("/balance", balanceHandler.GetBalance)
		api.POST("/balance/withdraw", balanceHandler.Withdraw)
	}
}

func SetupWithdrawalRoutes(r *gin.Engine, withdrawalHandler *handlers.WithdrawalHandler, authMW gin.HandlerFunc) {
	api := r.Group("/api/user")
	api.Use(authMW)

	api.GET("/withdrawals", authMW, withdrawalHandler.GetWithdrawals)

}
