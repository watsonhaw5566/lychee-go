package controller

import (
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
	response.SuccessWithData(c, "Welcome to Lychee-Go Framework!")
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
