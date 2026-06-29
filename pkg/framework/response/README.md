# response · 统一响应

标准化的 JSON 响应格式，避免每个接口自行组装。

## 用法

```go
import "lychee-go/internal/response"

// ---------- 成功响应 ----------
response.Success(c)                      // { code:0, message:"ok", data:null }
response.Success(c, user)                // { code:0, message:"ok", data:{...} }
response.OK(c, map[string]interface{}{   // 自定义返回数据
    "token": token,
    "user":  user,
})

// ---------- 分页响应 ----------
response.Page(c, userList, total, page, pageSize)
// { code:0, message:"ok", data:{list:[...], total:100, page:1, page_size:20} }

// ---------- 失败响应 ----------
response.Error(c, 400, "参数错误")
response.Error(c, 500, "服务器内部错误")
response.NotFound(c, "用户不存在")
response.Unauthorized(c, "请先登录")
response.Forbidden(c, "没有权限")
response.BadRequest(c, "邮箱格式不正确")
```

## 统一格式

```json
{
    "code": 0,
    "message": "ok",
    "data": { ... }
}
```

| code | 含义 |
|------|------|
| `0` | 成功 |
| `> 0` | 业务错误（前端按 code 分支处理） |
| HTTP 状态码 | `Error()` 直接透传 |

## Gin 集成示例

```go
r.POST("/api/login", func(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "参数格式错误")
        return
    }
    token, user := doLogin(req)
    response.OK(c, gin.H{"token": token, "user": user})
})
```
