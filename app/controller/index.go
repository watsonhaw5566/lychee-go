package controller

import (
	"lychee-go/internal/config"
	"lychee-go/internal/response"
	"time"

	"github.com/gin-gonic/gin"
)

type IndexController struct{}

func NewIndexController() *IndexController {
	return &IndexController{}
}

// Index 首页
func (ctrl *IndexController) Index(c *gin.Context) {
	response.Raw(c, gin.H{
		"message": "Welcome to Lychee-Go Framework!",
		"app": gin.H{
			"name":    config.GetString("app.name", "lychee-go"),
			"version": config.GetString("app.version", "1.0.0"),
			"debug":   config.GetBool("app.debug", true),
		},
		"docs": gin.H{
			"api":    "/api/users",
			"health": "/health",
		},
	})
}

// Health 健康检查
func (ctrl *IndexController) Health(c *gin.Context) {
	response.SuccessWithData(c, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
	})
}

// Ping ping 检查
func (ctrl *IndexController) Ping(c *gin.Context) {
	response.SuccessWithData(c, "pong")
}
