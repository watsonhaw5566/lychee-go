package middleware

import (
	"time"

	"lychee-go/internal/logger"

	"github.com/gin-gonic/gin"
)

// RequestLogger 请求日志中间件
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)

		if query != "" {
			logger.With(
				"method", c.Request.Method,
				"path", path,
				"query", query,
				"status", c.Writer.Status(),
				"latency", latency,
				"client_ip", c.ClientIP(),
			).Info("Request completed")
		} else {
			logger.With(
				"method", c.Request.Method,
				"path", path,
				"status", c.Writer.Status(),
				"latency", latency,
				"client_ip", c.ClientIP(),
			).Info("Request completed")
		}
	}
}
