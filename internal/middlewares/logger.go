package middlewares

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gophermart/pkg/logger"
)

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Log.Info("Incoming request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
		)
		c.Next()
	}
}
