# cors · 跨域资源共享

跨域中间件，支持 Origin 白名单、通配符、Preflight 预检请求、Vary 头。

## 用法

### 方式 1：使用 config.yml 的配置（推荐）

```go
import (
    "lychee-go/internal/cors"
    "github.com/gin-gonic/gin"
)

r := gin.Default()
r.Use(cors.Middleware())
```

`config.yml`：
```yaml
cors:
  allow_origins: "*"
  allow_methods: "GET,POST,PUT,DELETE,OPTIONS,PATCH"
  allow_headers: "Content-Type,Authorization,X-Requested-With"
  expose_headers: "X-RateLimit-Limit,X-RateLimit-Remaining,X-RateLimit-Reset"
  allow_credentials: true
  max_age: 86400
```

### 方式 2：开发环境完全放行

```go
r.Use(cors.AllowAll())
```

### 方式 3：特定域名白名单

```go
r.Use(cors.Restrict(
    "https://app.example.com",
    "https://admin.example.com",
    "https://cdn.example.com",
))
```

## 响应头说明

| 响应头 | 含义 |
|--------|------|
| `Access-Control-Allow-Origin` | 允许的来源域名 |
| `Access-Control-Allow-Methods` | 允许的 HTTP 方法 |
| `Access-Control-Allow-Headers` | 允许的请求头 |
| `Access-Control-Expose-Headers` | 前端可以读取的响应头 |
| `Access-Control-Allow-Credentials` | 是否允许携带 Cookie |
| `Access-Control-Max-Age` | Preflight 结果缓存时间（秒） |
| `Vary` | `Origin, Access-Control-Request-Method, Access-Control-Request-Headers` |

## Preflight 预检

浏览器对某些跨域请求会先发送 `OPTIONS` 预检请求。本中间件会：
1. 拦截所有 `OPTIONS` 请求
2. 检查 Origin 是否在白名单
3. 返回上述 CORS 头 + `HTTP 204 No Content`
4. 终止请求链（不再进入业务 Handler）

## 安全建议

**生产环境：**
- ❌ 不要 `allow_origins: "*"`
- ✅ 指定具体的域名（如 `https://app.example.com`）
- ✅ `allow_credentials: true`
- ✅ `expose_headers` 只列出前端确实需要读取的头

**开发环境：**
- 可以 `AllowAll()` 方便本地调试
