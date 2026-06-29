package route

import (
	"github.com/gin-gonic/gin"
)

// ResourceController 资源控制器接口
// 实现此接口的控制器可以使用 Resource 函数注册标准 CRUD 路由
type ResourceController interface {
	Index(c *gin.Context)  // GET /resource - 列表
	Show(c *gin.Context)   // GET /resource/:id - 详情
	Create(c *gin.Context) // POST /resource - 创建
	Update(c *gin.Context) // PUT /resource/:id - 更新
	Delete(c *gin.Context) // DELETE /resource/:id - 删除
}

// ResourceOptions 资源路由选项
type ResourceOptions struct {
	AuthMiddleware gin.HandlerFunc // 认证中间件
	Only           []string        // 只注册指定的方法: "index", "show", "create", "update", "delete"
	Except         []string        // 排除指定的方法
}

// Resource 注册标准资源路由
// path: 资源路径，如 "/users"
// ctrl: 实现了 ResourceController 接口的控制器
func Resource(g *gin.RouterGroup, path string, ctrl ResourceController, opts ...ResourceOptions) {
	var authMiddleware gin.HandlerFunc
	onlyMap := make(map[string]bool)
	exceptMap := make(map[string]bool)

	if len(opts) > 0 {
		authMiddleware = opts[0].AuthMiddleware
		for _, method := range opts[0].Only {
			onlyMap[method] = true
		}
		for _, method := range opts[0].Except {
			exceptMap[method] = true
		}
	}

	shouldRegister := func(method string) bool {
		if len(onlyMap) > 0 && !onlyMap[method] {
			return false
		}
		if exceptMap[method] {
			return false
		}
		return true
	}

	if shouldRegister("index") {
		g.GET(path, ctrl.Index)
	}

	if shouldRegister("show") {
		g.GET(path+"/:id", ctrl.Show)
	}

	if shouldRegister("create") {
		g.POST(path, ctrl.Create)
	}

	if shouldRegister("update") {
		if authMiddleware != nil {
			g.PUT(path+"/:id", authMiddleware, ctrl.Update)
		} else {
			g.PUT(path+"/:id", ctrl.Update)
		}
	}

	if shouldRegister("delete") {
		if authMiddleware != nil {
			g.DELETE(path+"/:id", authMiddleware, ctrl.Delete)
		} else {
			g.DELETE(path+"/:id", ctrl.Delete)
		}
	}
}
