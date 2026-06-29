# boot · 启动引导

框架启动时调用的模块，负责按正确顺序初始化其他模块。

## 核心行为：按需初始化

`boot.Boot()` 会读取 `config.yml`，**只有实际配置了的模块才会被初始化**。

- 未配置的模块 → 被跳过（SKIPPED），不浪费资源
- 已配置但初始化失败 → 输出 WARN 日志，标记 FAILED
- 已配置且初始化成功 → 输出 OK

> 这种设计的好处：你的 `config.yml` 写了什么，框架就用什么；没写的东西，框架不会尝试去连（比如没有 Redis 就不会去连 Redis，不会在启动日志里留下无关的失败信息）。

## 用法

```go
package main

import "lychee-go/internal/boot"

func main() {
    // 传入配置文件所在的目录
    if err := boot.Boot("config"); err != nil {
        panic(err)
    }

    // 之后就可以使用已配置好的各个模块
}
```

启动日志示例（以当前项目的 `config.yml` 为例）：

```
============================================================
Lychee-Go Framework boot completed!
Logger:     OK (always-on)
  - Database     OK (mysql@127.0.0.1:3306)
  - Cache        OK (redis@127.0.0.1:6379)
  - Filesystem   OK (driver: local)
  - JWT          OK (ttl: 86400s, max_per_user: 10)
  - Session      OK (ttl: 7200s)
  - Queue        OK (driver: memory)
  - Cron         OK (0 tasks registered)
  - CORS         OK
  - Throttle     OK (limit: 60/60s)
  - Cookie       OK (httponly: true, samesite: lax)
  - Swagger      OK (path: /swagger)
  - WebSocket    OK (read_buffer: 1024, write_buffer: 1024)
  - I18n         OK (default: zh-CN)
============================================================
```

如果删掉 `database` 块，Database 行就会变成 `SKIPPED (not configured)`。

---

## 模块初始化判定规则

下表说明每个模块在**什么情况下会被初始化**：

| 模块 | 判定条件 |
|------|---------|
| **Logger** | 无条件（基础模块，必启动） |
| **Database** | `database.host` **且** `database.database` 同时非空 |
| **Cache** | `cache.driver` **且** `cache.host` 非空，且 driver 不是 `none` / `disabled` |
| **Filesystem** | `filesystem.default` 非空，且不是 `none` / `disabled` |
| **JWT (auth)** | `auth.jwt_secret` 非空 |
| **Session** | `session.ttl > 0`（或显式 `session.enabled: true`） |
| **Queue** | `queue.driver` 非空，且不是 `none` / `disabled` |
| **Cron** | 默认启用；可通过 `cron.enabled: false` 显式禁用 |
| **CORS** | `cors.allow_origins` 非空，且不是 `none` / `disabled` |
| **Throttle** | `throttle.limit > 0` **且** `throttle.window > 0`（或显式 `throttle.enabled: true`） |
| **Cookie** | `cookie.secret` 非空 |
| **Swagger** | 默认启用；可通过 `swagger.enabled: false` 显式禁用 |
| **WebSocket** | 默认启用；可通过 `websocket.enabled: false` 显式禁用 |
| **I18n** | 默认启用；可通过 `i18n.enabled: false` 显式禁用 |

### 显式禁用的写法

如果你想保留配置结构但暂时禁用某个模块，可以用以下"关闭写法"：

```yaml
cache:
  driver: none      # ← 禁用缓存

cors:
  allow_origins: none  # ← 禁用跨域

cron:
  enabled: false    # ← 禁用定时任务

throttle:
  enabled: false    # ← 禁限流
```

> 这种方式比直接删掉整块更清晰：你一眼就能看出"我考虑过这个模块，只是当前不需要"。

---

## 初始化顺序

启动的严格顺序（依赖在前的先初始化）：

```
1. config     ← 必须第一个，所有模块都依赖它
2. logger     ← 必须第二个，所有模块都用它输出日志
3. database   ← 可选
4. cache      ← 可选（其他模块如 queue 可能依赖它）
5. filesystem ← 可选
6. jwt        ← 可选（需要 auth.jwt_secret）
7. session    ← 可选
8. queue      ← 可选（可能依赖 cache）
9. cron       ← 可选（纯内存，默认启用）
10. cors      ← 可选
11. throttle  ← 可选
12. cookie    ← 可选（需要 cookie.secret）
13. swagger   ← 可选（生成 API 文档）
14. websocket ← 可选（WebSocket 服务）
15. i18n      ← 可选（国际化支持）
```

---

## 与 `internal/config` 新增 API

`boot.go` 的判断基于以下三个新增的配置检测方法：

| 方法 | 说明 |
|------|------|
| `config.IsSet("database.host")` | 配置文件中是否存在该 key |
| `config.HasSection("database")` | 配置文件中是否存在该 section 块 |
| `config.IsConfigured("database.host")` | 该 key 存在 **且** 值非空（去除空白） |

如果你想在业务代码中做类似的"配置开关"判断，这三个函数都可以直接使用。

### 在业务中使用示例

```go
if config.IsConfigured("database.host") {
    // 只有配了数据库才执行这段逻辑
    users, _ := db.Select("users")
}

if config.HasSection("auth") {
    // 只有配了鉴权模块才启用登录接口
    r.POST("/login", controller.Login)
}
```

---

## 设计要点

1. **幂等性**：所有模块的 `Init()` 内部都使用了 `sync.Once`，多次调用 `boot.Boot()` 不会重复初始化。
2. **失败隔离**：单个模块初始化失败（如 MySQL 连不上）只会影响自己，不会让整个服务起不来。
3. **零配置可用**：即使 `config.yml` 中只保留 `app` 和 `log` 两个块，框架依然可以正常启动。
4. **依赖安全**：如果 `cache` 未配置，`queue` 会自动选择 `memory` 驱动（不会因为缺 Redis 而挂掉）。