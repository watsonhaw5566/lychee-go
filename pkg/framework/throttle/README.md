# throttle · 接口限流

令牌桶算法 + 滑动窗口的接口限流，支持按 IP / 用户 / 路由 / 全局 多维度限流。

## 基本用法

```go
import (
    "lychee-go/internal/throttle"
    "github.com/gin-gonic/gin"
)

r := gin.Default()

// 方式 1：使用 config.yml 配置（默认按 IP 限流，每分钟 60 次）
r.Use(throttle.Middleware())

// 方式 2：预定义常用配置
api := r.Group("/api")
api.Use(throttle.API())                    // 每分钟 60 次 / IP

login := r.Group("/login")
login.POST("", throttle.Middleware(throttle.Login()), loginHandler)
// Login 配置：每 5 分钟 10 次 / IP（防止暴力破解）

// 方式 3：按时间粒度
r.GET("/p1", throttle.PerMinute(100), ...)  // 每 IP 每分钟 100 次
r.GET("/p2", throttle.PerSecond(10),  ...)  // 每 IP 每秒 10 次
r.GET("/p3", throttle.PerHour(1000),  ...)  // 每 IP 每小时 1000 次

// 方式 4：全局 QPS 限制（所有 IP 共享一个桶）
r.Use(throttle.Global(1000, time.Second))   // 全服务器每秒最多 1000 请求

// 方式 5：按路由限流
r.POST("/order", throttle.PerRoute(30, time.Minute), orderHandler)

// 方式 6：按用户限流（需要 header X-User-ID 或 gin context user_id）
r.GET("/user/profile", throttle.PerUser(30, time.Minute), profileHandler)
```

## 配置

```yaml
throttle:
  limit: 60        # 每个时间窗口最大请求数
  window: 60       # 时间窗口（秒）
  prefix: "lychee_go_"
```

## 超限响应

HTTP 状态码 **`429 Too Many Requests`**

```json
{
    "code": 429,
    "message": "请求过于频繁，请稍后再试",
    "retry_after_seconds": 45,
    "reset_at": "2026-06-29T12:34:56+08:00"
}
```

响应头包含标准限流信息：
```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1750000123
Retry-After: 45
```

## 手动调用（非中间件场景）

```go
// 短信发送限流：每个手机号 1 小时最多 5 条
key := "sms:" + phone
allowed, remaining, resetAt := throttle.Allow(key, 5, time.Hour)
if !allowed {
    return errors.New("短信发送过于频繁，请稍后再试")
}
sms.Send(phone, code)

// 手动重置
throttle.Reset(key)
```

## 存储

默认使用 **内存存储**（`MemoryStore`），适合单机部署。

### 扩展到 Redis（分布式）

实现 `throttle.Store` 接口：

```go
type Store interface {
    Take(key string, limit int, window time.Duration) (bool, int, time.Time)
    Reset(key string)
    Stats(key string, limit int, window time.Duration) (int, time.Time)
}

// 注册
throttle.SetStore(&RedisStore{...})
```

## 真实 IP 检测

自动从以下 header 读取真实客户端 IP：
1. `X-Forwarded-For`（取第一个）
2. `X-Real-IP`
3. Gin 的 `ClientIP()`（回退）

这样即使部署在 Nginx / CDN 后面，也能正确按客户端 IP 限流。
