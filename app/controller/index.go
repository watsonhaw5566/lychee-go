package controller

import (
	"lychee-go/internal/response"
	"lychee-go/internal/view"
	"time"

	"github.com/gin-gonic/gin"
)

type IndexController struct{}

func NewIndexController() *IndexController {
	return &IndexController{}
}

func (ctrl *IndexController) Index(c *gin.Context) {
	data := map[string]interface{}{
		"Title":   "Lychee-Go Framework",
		"Version": "1.0.0",
		"Message": "Welcome to Lychee-Go Framework!",
	}
	view.Render(c.Writer, "index", data)
}

func (ctrl *IndexController) Health(c *gin.Context) {
	response.SuccessWithData(c, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
	})
}

func (ctrl *IndexController) Ping(c *gin.Context) {
	response.SuccessWithData(c, "pong")
}
