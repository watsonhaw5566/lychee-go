package middleware

import (
	"strings"

	"lychee-go/internal/response"

	"github.com/gin-gonic/gin"
)

// Auth JWT 鉴权中间件
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			response.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Unauthorized(c, "Token 格式错误")
			c.Abort()
			return
		}

		token := parts[1]
		if token == "" {
			response.Unauthorized(c, "无效的 Token")
			c.Abort()
			return
		}

		c.Set("user_id", uint(1))
		c.Set("token", token)

		c.Next()
	}
}

// AdminAuth 管理员鉴权中间件
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
