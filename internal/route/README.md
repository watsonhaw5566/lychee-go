# route · 资源路由封装

提供类似 Rails 的资源路由功能，支持标准 CRUD 操作的自动注册。

## 使用方式

```go
import internalroute "lychee-go/internal/route"

// 标准用法
internalroute.Resource(api, "/users", userCtrl, internalroute.ResourceOptions{
    AuthMiddleware: middleware.Auth(),
})

// 只注册部分方法
internalroute.Resource(api, "/posts", postCtrl, internalroute.ResourceOptions{
    AuthMiddleware: middleware.Auth(),
    Only: []string{"index", "show"},
})

// 排除某些方法
internalroute.Resource(api, "/comments", commentCtrl, internalroute.ResourceOptions{
    AuthMiddleware: middleware.Auth(),
    Except: []string{"delete"},
})

// 不需要认证的资源
internalroute.Resource(api, "/public", publicCtrl)
```

## 路由映射

| HTTP 方法 | 路径 | 控制器方法 |
|-----------|------|------------|
| GET | /resource | Index |
| GET | /resource/:id | Show |
| POST | /resource | Create |
| PUT | /resource/:id | Update |
| DELETE | /resource/:id | Delete |

## 接口定义

控制器需要实现 `ResourceController` 接口：

```go
type ResourceController interface {
    Index(c *gin.Context)
    Show(c *gin.Context)
    Create(c *gin.Context)
    Update(c *gin.Context)
    Delete(c *gin.Context)
}
```