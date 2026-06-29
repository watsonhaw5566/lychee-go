package middleware

import (
	"lychee-go/internal/config"

	"github.com/gin-gonic/gin"
)

// CORS 跨域处理中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		allowOrigins := config.GetString("cors.allow_origins", "*")
		allowMethods := config.GetString("cors.allow_methods", "GET,POST,PUT,DELETE,OPTIONS")
		allowHeaders := config.GetString("cors.allow_headers", "Content-Type,Authorization")
		allowCredentials := config.GetBool("cors.allow_credentials", true)

		c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigins)
		c.Writer.Header().Set("Access-Control-Allow-Methods", allowMethods)
		c.Writer.Header().Set("Access-Control-Allow-Headers", allowHeaders)

		if allowCredentials {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
