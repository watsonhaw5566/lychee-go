# Lychee-Go

> 🍒 一款轻量级的 Go Web 框架 —— 灵感来自 ThinkPHP，拥抱 Gin 的性能与简洁。

---

## ⚠️ 重要提示：**开发中 · 生产未就绪**

> **请不要在生产环境中使用本项目。**

**Lychee-Go** 目前仍处于 **早期开发 / 学习练手阶段**，存在以下已知问题，不适合承载任何线上业务：

| 风险点 | 现状 | 影响 |
|--------|------|------|
| 🧪 **未经生产验证** | 没有在任何真实线上业务跑过，也没有经过压力测试 / 渗透测试 | ❌ **高** |
| 🔒 **安全未审计** | 签名算法、JWT、Session、CORS 等关键代码未做安全审计；默认密钥在 config.yml 中明文可见 | ❌ **高** |
| 🐛 **可能存在 Bug** | DB / Cache / Queue 等模块使用大量自实现代码，缺少单元测试覆盖 | ⚠️ **中** |
| 🚧 **API 可能变化** | 接口名 / 返回结构 / 模块设计随时可能调整；不提供向后兼容承诺 | ⚠️ **中** |
| 📚 **文档不完整** | 各模块 README 只是使用示例，不是正式 API 文档 | ⚠️ **中** |
| 📦 **无稳定发布** | 没有打 tag / 没有版本分支 / 没有 changelog | 🟡 **低** |

> **推荐用途**：学习 Go / 学习 Web 框架设计 / 个人小项目 / 技术调研。
> **请勿用于**：生产环境 / 商用系统 / 涉及敏感数据的任何场景。
>
> 如果你在寻找**可直接用于生产**的 Go 框架，推荐直接使用：
> - **Gin** —— 久经沙场的高性能 HTTP 框架（Lychee-Go 底层就是 Gin）
> - **Go-Zero / Kratos** —— 带代码生成、服务治理的微服务框架

---

## 🌱 项目简介

**Lychee-Go** 是一款**面向学习与小项目**的轻量级 Go Web 框架，致力于提供一套**开箱即用、零魔法、可追踪**的基础组件。

> 💡 **设计哲学**：单一配置文件 · 模块化架构 · 约定优于配置 · 拒绝黑盒魔法
>
> 💡 **定位**：学习练手的框架骨架，供你**在此基础上自己改、自己加、自己打磨**。

### ✨ 核心特性（当前能力）

| | 特性 | 说明 |
|--|------|------|
| ⚡ | **原生高性能** | 基于 Gin 封装，零额外抽象层开销，享受 Go 原生并发性能 |
| 📦 | **单体配置** | 所有 15+ 个模块共享一份 `config.yml`，类似 Spring Boot 的体验 |
| 🧩 | **模块化架构** | 每个模块都有 README，按需启用、互不耦合（未配置的模块自动跳过） |
| 🔌 | **WebSocket 支持** | 内置 WebSocket 服务，支持消息广播、客户端管理、自定义消息处理器 |
ei'w| 🌍 | **国际化支持** | 内置 i18n 模块，支持多语言翻译、自动语言检测 |
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
# 克隆（如已有项目，跳过）
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

# ...其他模块的配置（见 config.yml 原文件）
```

> 没有 MySQL/Redis 也能启动，框架会跳过失败的初始化并输出 `Warn` 日志。

### 3. 启动

```bash
go run main.go
```

输出示例：

```
============================================================
Lychee-Go Framework starting...
============================================================
App Name: lychee-go, Version: 1.0.0
Debug Mode: true, Port: 8080
============================================================
Lychee-Go Framework boot completed!
  - Logger:     OK
  - Filesystem: OK
  - JWT:        OK
  - Session:    OK
  - Queue:      OK
  - Cron:       OK
  - CORS:       OK
  - Throttle:   OK
  - Cookie:     OK
============================================================
⇨ http server started on [::]:8080
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
    // 全局中间件
    r.Use(cors.Middleware())       // 跨域
    r.Use(throttle.PerMinute(60))   // 每 IP 每分钟 60 次

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
    "lychee-go/internal/response"
    "lychee-go/internal/validation"
    "lychee-go/internal/jwt"

    "github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    c.ShouldBindJSON(&req)

    // 参数验证
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

    // 验证账号...
    userID := uint(123)

    // 签发 Token
    token, _, _ := jwt.Login(userID, "api", nil)
    response.OK(c, gin.H{"token": token, "user_id": userID})
}
```

就是这么简单。

---

## 🧱 模块索引（18 个）

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
| [**jwt**](internal/jwt/README.md) | 轻量 Token 鉴权（类似 Sa-Token） | — |
| [**satoken**](internal/satoken/README.md) | Sa-Token 权限认证框架（登录/登出/续期/踢人） | `go-redis/v9` |
| [**validation**](internal/validation/README.md) | 参数验证（20+ 规则 + 自定义） | — |
| [**filesystem**](internal/filesystem/README.md) | 文件系统（多驱动接口） | — |
| [**queue**](internal/queue/README.md) | 消息队列（Redis/Memory） | `go-redis/v9` |
| [**cron**](internal/cron/README.md) | 定时任务（6 字段 Cron） | — |
| [**swagger**](internal/swagger/README.md) | API 文档生成 | — |
| [**websocket**](internal/websocket/README.md) | WebSocket 服务（广播/消息处理） | `gorilla/websocket` |

**公用工具有 2 个**（可在业务中随意引用）：

| 模块 | 核心功能 |
|------|---------|
| `pkg/helper` | 字符串 / 数组 / Map 工具函数 |
| `pkg/utils`  | MD5 / SHA / Bcrypt Hash |

---

## 🔧 常见场景速查

### 场景 1：参数校验

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

### 场景 2：登录鉴权（JWT）

```go
import "lychee-go/internal/jwt"

// 登录：签发 Token
token, _, _ := jwt.Login(userID, "api", map[string]interface{}{
    "username": "alice",
})

// 接口：验证 Token
claims, err := jwt.Verify(tokenFromHeader)
userID := claims.UserID

// 登出 / 强制下线
jwt.Logout(token)
jwt.KickOut(userID)
```

### 场景 3：签名 Cookie

```go
import "lychee-go/internal/cookie"

// 存（防篡改）
cookie.SetSigned(c, "user_id", "123")

// 读
userID, _ := cookie.GetSigned(c, "user_id")
```

### 场景 4：异步队列

```go
import "lychee-go/internal/queue"

// 投递任务
queue.Dispatch("default", "send_email", map[string]interface{}{
    "to": "user@example.com",
    "subject": "欢迎注册",
})

// main 中启动 Worker
go queue.StartWorker("default")
```

### 场景 5：定时任务

```go
import "lychee-go/internal/cron"

// 每天凌晨 2 点清理日志
cron.AddFunc("clean_logs", "0 0 2 * * *", func() error {
    logger.Info("开始清理日志...")
    return nil
})
```

### 场景 6：接口限流

```go
import "lychee-go/internal/throttle"

// 登录接口：5 分钟最多 10 次（防暴力破解）
r.POST("/login", throttle.Middleware(throttle.Login()), loginHandler)

// API 通用：每分钟 60 次
api.Use(throttle.API())
```

### 场景 7：WebSocket 实时通信

```go
import "lychee-go/internal/websocket"

// 注册消息处理器
websocket.RegisterHandler("chat", func(conn *websocket.Conn, msg websocket.Message) error {
    var data struct {
        User    string `json:"user"`
        Content string `json:"content"`
    }
    json.Unmarshal(msg.Payload, &data)
    
    // 广播消息给所有在线客户端
    return websocket.Broadcast("chat", data)
})

// 配置 WebSocket 路由
r.GET("/ws", websocket.HandleWebSocket)

// 在任意地方推送通知
websocket.Broadcast("notification", map[string]string{
    "title": "系统通知",
    "body":  "您有新消息",
})
```

---

## 🔒 安全默认值

框架内置了以下安全策略（可在 `config.yml` 中覆盖）：

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

- **业务代码** 放在 `app/`（或你习惯的其他目录）
- **框架代码** 全部在 `internal/`（Go 编译器会阻止外部项目引用）
- **通用工具** 放在 `pkg/`（业务代码和框架代码都可引用）

---

## 🛣️ 生产部署建议

1. **配置文件分离**：`config.yml` 不要放到容器镜像中，以 `--config` 参数或 Volume 挂载方式注入。
2. **日志**：生产环境设置 `log.level: info`，文件大小/备份数按需调整。
3. **Redis**：限流、队列、缓存建议统一使用 Redis 存储（多实例部署必备）。
4. **CORS**：生产环境必须用具体域名（`allow_origins: https://app.example.com`），禁止 `*`。
5. **Cookie**：生产环境 `secure: true`，`samesite: strict`。
6. **密钥**：`jwt.secret` 和 `cookie.secret` 请使用 **不同的、强随机的** 密钥，不要提交到 Git。

---

## 🎯 为什么选择 Lychee-Go？

| 对比维度 | Lychee-Go | 裸 Gin | 其他 Go 框架 |
|---------|-----------|--------|--------------|
| **上手难度** | ⭐ 极低（ThinkPHP 风格 API） | 中（需自己搭基建） | 高（学习曲线陡峭） |
| **配置文件** | 单一 `config.yml` | 无（自行实现） | 多文件 / DSL |
| **内置模块** | 15 个开箱即用 | 仅路由 + 中间件 | 各有取舍 |
| **代码可读性** | 每个模块独立 README | 需阅读第三方文档 | 文档分散 |
| **可追踪** | 零黑盒魔法，调用链清晰 | 清晰 | 依赖重、抽象多 |
| **生产就绪** | ✅ 安全默认值齐全 | ❌ 需自行补齐 | ✅ |

---

## 📖 学习路径建议

无论你是新手还是从 PHP 迁移过来，都可以按以下顺序上手：

1. **第 1 步 · 理解启动**：阅读 `main.go` + `internal/boot/boot.go`，明白框架是如何一步步唤醒所有模块的。
2. **第 2 步 · 写第一个 API**：照着 `app/controller/user.go` 的样子，写一个属于你的接口。
3. **第 3 步 · 翻阅模块文档**：打开 `internal/` 下任意模块的 README，每个都是独立可阅读的小文档。
4. **第 4 步 · 动手修改**：试着换一个驱动（比如 session 从内存切到 Redis），或加一个自定义的验证规则。

> 🍀 框架的每个模块都独立且足够小，改起来不怕坏 —— 这正是它作为学习项目的价值所在。

---

## 📄 License

MIT — 欢迎学习、复制、修改、二次创作。