package route

import (
	"lychee-go/app/controller"
	"lychee-go/app/middleware"

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

	// ======== API v1 路由组 ========
	api := r.Group("/api")
	{
		// 用户相关（公开）
		userCtrl := controller.NewUserController()
		api.GET("/users", userCtrl.GetList)
		api.GET("/users/:id", userCtrl.GetUser)
		api.POST("/users", userCtrl.CreateUser)

		// 需要登录的路由
		auth := api.Group("")
		auth.Use(middleware.Auth())
		{
			auth.PUT("/users/:id", userCtrl.UpdateUser)
			auth.DELETE("/users/:id", userCtrl.DeleteUser)
		}
	}
}
