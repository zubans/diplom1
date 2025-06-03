package routes

import (
	"github.com/gin-gonic/gin"
	"gophermart/internal/handlers"
)

func SetupUserRoutes(r *gin.Engine, userHandler *handlers.Server) {
	api := r.Group("/api/user")
	{
		api.POST("/register", userHandler.Register)
		api.POST("/login", userHandler.Login)
	}
}
