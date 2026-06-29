# Lychee-Go

> 🍒 一款轻量级的 Go Web 框架 —— 灵感来自 ThinkPHP，拥抱 Gin 的性能与简洁。

---

## 🚀 积极开发中 · 生产未就绪

**Lychee-Go** 正在积极开发中，欢迎参与贡献！

> **⚠️ 注意**：项目目前处于快速迭代阶段，**暂不建议用于生产环境**。API 可能会有 Breaking Change，安全审计尚未完成。

当前已知的改进方向：
- ✅ 核心模块功能完善
- ⚠️ 安全审计进行中
- 🚧 单元测试覆盖
- 📝 文档完善

---

## 🌱 项目简介

**Lychee-Go** 是一款轻量级的 Go Web 框架，致力于提供一套**开箱即用、零魔法、可追踪**的基础组件。

> 💡 **设计哲学**：单一配置文件 · 模块化架构 · 约定优于配置 · 拒绝黑盒魔法

### ✨ 核心特性

| | 特性 | 说明 |
|--|------|------|
| ⚡ | **原生高性能** | 基于 Gin 封装，零额外抽象层开销，享受 Go 原生并发性能 |
| 📦 | **单体配置** | 所有 15+ 个模块共享一份 `config.yml`，类似 Spring Boot 的体验 |
| 🧩 | **模块化架构** | 每个模块都有 README，按需启用、互不耦合（未配置的模块自动跳过） |
| 🔌 | **WebSocket 支持** | 内置 WebSocket 服务，支持消息广播、客户端管理、自定义消息处理器 |
| 🌍 | **国际化支持** | 内置 i18n 模块，支持多语言翻译、自动语言检测 |
| 🔐 | **全链路安全** | 内置签名 Cookie / JWT Token / 接口限流 / CORS 防护 / Session 管理 |
| 🧪 | **零依赖启动** | 没有 MySQL 或 Redis 也能正常启动（失败模块仅 Warn，不阻塞） |
| 🧠 | **ThinkPHP 友好** | API 风格贴近 ThinkPHP，PHP 开发者可无痛迁移到 Go |
| 🚀 | **完整基建** | 参数验证 / 消息队列 / 定时任务 / 文件系统 / 统一响应 / Swagger / 命令行 —— 全部内置 |
| 📖 | **自文档化** | 每个模块自带 README，代码即文档 |
| 🔥 | **热重载** | 内置 `.air.toml` 配置，`air` 启动即可文件变化自动编译重启 |

---

## 📂 目录结构

```
lychee-go/
├── main.go                 # 程序入口
├── go.mod / go.sum         # 依赖管理
├── config/
│   └── config.yml          # 单一配置文件（类似 Spring Boot）
├── internal/               # 🔴 框架核心（16 个模块）
│   ├── boot/               # 启动引导
│   ├── config/             # 配置管理 (viper)
│   ├── logger/             # 日志系统 (zap)
│   ├── db/                 # 数据库 (GORM)
│   ├── cache/              # 缓存 (Redis)
│   ├── response/           # 统一响应
│   ├── cors/               # 跨域中间件
│   ├── throttle/           # 接口限流
│   ├── cookie/             # Cookie 管理（签名/JSON/Flash）
│   ├── session/            # 会话管理
│   ├── jwt/                # Token 鉴权
│   ├── satoken/            # Sa-Token 权限认证框架
│   ├── validation/         # 参数验证（20+ 规则）
│   ├── filesystem/         # 文件系统（多驱动）
│   ├── queue/              # 消息队列（Redis/Memory）
│   ├── cron/               # 定时任务（6 字段 Cron）
│   ├── swagger/            # API 文档生成
│   └── websocket/          # WebSocket 服务
├── app/                    # 🟢 应用代码（你写的业务）
│   ├── controller/
│   ├── model/
│   ├── service/
│   ├── middleware/
│   └── route/
├── pkg/                    # 🟢 公开可复用工具包
│   ├── helper/             # 字符串 / 数组工具
│   └── utils/              # 加密 / Hash
└── runtime/                # 运行时（日志 / 上传文件）
    └── .gitkeep
```

> 💡 **`internal/`** 目录下的每个模块都有一份独立的 `README.md`，打开即可查阅详细用法。

---

## 🚀 快速开始

### 1. 安装

```bash
cd lychee-go
go mod download
```

### 2. 配置

编辑 `config/config.yml`，根据实际环境修改：

```yaml
app:
  name: lychee-go
  port: 8080
  debug: true

database:
  driver: mysql
  host: 127.0.0.1
  port: 3306
  database: lychee_go
  username: root
  password: "123456"

cache:
  host: 127.0.0.1
  port: 6379
  prefix: lychee_go_
```

> 没有 MySQL/Redis 也能启动，框架会跳过失败的初始化并输出 `Warn` 日志。

### 3. 启动

```bash
go run main.go
```

浏览器访问 `http://localhost:8080/ping`，预期返回：

```json
{"code":0,"message":"ok","data":"pong"}
```

---

## 👋 一个最小例子：用户 API

在 `app/controller/user.go` 中定义 Handler，在 `app/route/route.go` 注册路由：

```go
// app/route/route.go
package route

import (
    "lychee-go/app/controller"
    "lychee-go/internal/cors"
    "lychee-go/internal/throttle"
    "github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
    r.Use(cors.Middleware())
    r.Use(throttle.PerMinute(60))

    api := r.Group("/api")
    {
        api.POST("/login", controller.Login)
        api.GET("/users", controller.UserList)
        api.GET("/users/:id", controller.UserDetail)
    }
}
```

在 `app/controller/user.go` 中：

```go
package controller

import (
    "lychee-go/internal/jwt"
    "lychee-go/internal/response"
    "lychee-go/internal/validation"
    "github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    c.ShouldBindJSON(&req)

    v := validation.New()
    v.WithRule("email", "required|email")
    v.WithRule("password", "required|min:6")
    if ok, errors := v.Validate(map[string]interface{}{
        "email":    req.Email,
        "password": req.Password,
    }); !ok {
        response.BadRequest(c, validation.FirstError(errors))
        return
    }

    userID := uint(123)
    token, _, _ := jwt.Login(userID, "api", nil)
    response.OK(c, gin.H{"token": token, "user_id": userID})
}
```

---

## 🧱 模块索引

每个模块的详细用法，请查看 `internal/模块名/README.md`。

| 模块 | 核心功能 | 外部依赖 |
|------|---------|---------|
| [**boot**](internal/boot/README.md) | 启动引导、初始化顺序 | — |
| [**config**](internal/config/README.md) | 配置加载（YAML + env） | `spf13/viper` |
| [**logger**](internal/logger/README.md) | 结构化日志 | `uber-go/zap`, `lumberjack` |
| [**db**](internal/db/README.md) | 数据库 ORM | `gorm.io/gorm` |
| [**cache**](internal/cache/README.md) | Redis 缓存 | `go-redis/v9` |
| [**response**](internal/response/README.md) | 统一 JSON 响应 | `gin-gonic/gin` |
| [**cors**](internal/cors/README.md) | CORS 跨域中间件 | `gin-gonic/gin` |
| [**throttle**](internal/throttle/README.md) | 令牌桶接口限流 | `gin-gonic/gin` |
| [**cookie**](internal/cookie/README.md) | Cookie 管理（签名/JSON/Flash） | `gin-gonic/gin` |
| [**session**](internal/session/README.md) | 会话管理（KV + 自动过期） | — |
| [**jwt**](internal/jwt/README.md) | 轻量 Token 鉴权 | — |
| [**satoken**](internal/satoken/README.md) | Sa-Token 权限认证框架 | `go-redis/v9` |
| [**validation**](internal/validation/README.md) | 参数验证（20+ 规则 + 自定义） | — |
| [**filesystem**](internal/filesystem/README.md) | 文件系统（多驱动接口） | — |
| [**queue**](internal/queue/README.md) | 消息队列（Redis/Memory） | `go-redis/v9` |
| [**cron**](internal/cron/README.md) | 定时任务（6 字段 Cron） | — |
| [**swagger**](internal/swagger/README.md) | API 文档生成 | — |
| [**websocket**](internal/websocket/README.md) | WebSocket 服务（广播/消息处理） | `gorilla/websocket` |
| [**command**](internal/command/README.md) | 命令行命令注册与执行 | — |
| [**env**](internal/env/README.md) | 环境变量加载与管理 | — |
| [**i18n**](internal/i18n/README.md) | 国际化支持（多语言翻译） | — |

**公用工具有 2 个**：

| 模块 | 核心功能 |
|------|---------|
| `pkg/helper` | 字符串 / 数组 / Map 工具函数 |
| `pkg/utils`  | MD5 / SHA / Bcrypt Hash |

---

## 🔧 常见场景速查

### 参数校验

```go
import "lychee-go/internal/validation"

v := validation.New()
v.WithRule("email", "required|email")
v.WithRule("password", "required|password")
ok, errs := v.Validate(map[string]interface{}{
    "email": "alice@example.com",
    "password": "abc123",
})
```

### 登录鉴权（JWT）

```go
import "lychee-go/internal/jwt"

token, _, _ := jwt.Login(userID, "api", map[string]interface{}{
    "username": "alice",
})
claims, err := jwt.Verify(tokenFromHeader)
jwt.Logout(token)
jwt.KickOut(userID)
```

### 签名 Cookie

```go
import "lychee-go/internal/cookie"

cookie.SetSigned(c, "user_id", "123")
userID, _ := cookie.GetSigned(c, "user_id")
```

### 异步队列

```go
import "lychee-go/internal/queue"

queue.Dispatch("default", "send_email", map[string]interface{}{
    "to": "user@example.com",
    "subject": "欢迎注册",
})
go queue.StartWorker("default")
```

### 定时任务

```go
import "lychee-go/internal/cron"

cron.AddFunc("clean_logs", "0 0 2 * * *", func() error {
    logger.Info("开始清理日志...")
    return nil
})
```

### WebSocket 实时通信

```go
import "lychee-go/internal/websocket"

websocket.RegisterHandler("chat", func(conn *websocket.Conn, msg websocket.Message) error {
    return websocket.Broadcast("chat", data)
})
r.GET("/ws", websocket.HandleWebSocket)
```

---

## 🔒 安全默认值

| 策略 | 默认值 | 可配置 |
|------|--------|--------|
| Cookie `HttpOnly` | `true` | ✅ |
| Cookie `SameSite` | `lax` | ✅ |
| JWT 默认过期 | 24 小时 | ✅ |
| JWT 单用户并发 | 10 个 Token | ✅ |
| Session 过期 | 2 小时 | ✅ |
| Throttle 默认 | 60 次 / 分钟 · IP | ✅ |
| CORS 凭证 | `true` | ✅ |

---

## 🧩 框架与应用的分层

```
┌──────────────────────────────┐
│     app/ （你的业务代码）     │
│  controller / service / model │
└──────────────────────────────┘
                │ 使用
                ▼
┌──────────────────────────────┐
│   internal/ （框架核心）      │
│  15 个模块，提供基础能力      │
└──────────────────────────────┘
                │ 使用
                ▼
┌──────────────────────────────┐
│   pkg/ （公开工具）           │
│   helper / utils             │
└──────────────────────────────┘
                │ 使用
                ▼
┌──────────────────────────────┐
│   第三方库（Gin / GORM ...）  │
└──────────────────────────────┘
```

---

## 🛣️ 生产部署建议

1. **配置文件分离**：`config.yml` 不要放到容器镜像中，以 Volume 挂载方式注入。
2. **日志**：生产环境设置 `log.level: info`。
3. **Redis**：限流、队列、缓存建议统一使用 Redis 存储。
4. **CORS**：生产环境必须用具体域名，禁止 `*`。
5. **Cookie**：生产环境 `secure: true`，`samesite: strict`。
6. **密钥**：`jwt.secret` 和 `cookie.secret` 请使用强随机密钥，不要提交到 Git。

---

## 📄 License

MIT