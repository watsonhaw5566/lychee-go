package route

import (
	"lychee-go/app/controller"
	"lychee-go/app/middleware"
	internalroute "lychee-go/internal/route"

	"github.com/gin-gonic/gin"
)

// Register 注册所有路由
func Register(r *gin.Engine) {
	// ======== 全局中间件 ========
	r.Use(middleware.CORS())
	r.Use(middleware.RequestLogger())
	r.Use(gin.Recovery())

	// ======== 首页 / 健康检查 ========
	indexCtrl := controller.NewIndexController()
	r.GET("/", indexCtrl.Index)
	r.GET("/health", indexCtrl.Health)
	r.GET("/ping", indexCtrl.Ping)

	// ======== API v1 路由组 ========
	api := r.Group("/api")
	{
		// 用户资源路由（使用 internal/route 的 Resource 函数自动注册 CRUD 路由）
		userCtrl := controller.NewUserController()
		internalroute.Resource(api, "/users", userCtrl, internalroute.ResourceOptions{
			AuthMiddleware: middleware.Auth(),
		})
	}
}
