# cache · 缓存

支持多种驱动的缓存模块，包括内存、文件和 Redis，参考 think-cache 设计。

## 驱动类型

| 驱动 | 说明 | 适用场景 |
|------|------|----------|
| **memory** | 内存驱动 | 开发测试、小场景（进程退出后数据丢失） |
| **file** | 文件驱动 | 需要持久化但不需要 Redis 的场景 |
| **redis** | Redis 驱动 | 生产环境、分布式场景 |

## 配置

```yaml
cache:
  driver: memory           # 驱动类型: memory, file, redis
  prefix: lychee_go_      # 自动给每个 key 加前缀，避免多项目冲突
  default_ttl: 3600       # 默认 TTL（秒）
  
  # 文件驱动配置（driver 为 file 时生效）
  file:
    path: runtime/cache   # 缓存文件存储目录
  
  # Redis 驱动配置（driver 为 redis 时生效）
  host: 127.0.0.1
  port: 6379
  password: ""
  database: 0
```

## 用法

```go
import "lychee-go/internal/cache"

// ---------- 基本 KV ----------
cache.Set("user:123", userData, 3600)           // 存，TTL 单位秒
data, err := cache.Get("user:123")              // 取
cache.Delete("user:123")                        // 删
cache.Has("user:123")                           // 是否存在

// ---------- 自增自减 ----------
cache.Incr("counter:views")                     // +1
cache.IncrBy("counter:views", 10)               // +10
cache.Decr("counter:views")                     // -1
cache.DecrBy("counter:views", 5)                // -5

// ---------- 过期 ----------
cache.Expire("session:abc", 7200)               // 设置过期时间（秒）
cache.TTL("session:abc")                        // 获取剩余 TTL

// ---------- 分布式锁 ----------
cache.SetNX("lock:order", "1", 10)              // 只有不存在时才设置

// ---------- Hash 操作 ----------
cache.HSet("user:123", "name", "Alice")
name, _ := cache.HGet("user:123", "name")
user, _ := cache.HGetAll("user:123")

// ---------- List 操作 ----------
cache.LPush("queue:tasks", "task1", "task2")
task, _ := cache.RPop("queue:tasks")
len, _ := cache.LLen("queue:tasks")

// ---------- 缓存标签 ----------
cache.SetWithTags("user:123", userData, []string{"user", "user:123"}, 3600)
cache.Tag("user").Flush()                        // 删除标签下所有缓存
cache.FlushTags("user", "article")              // 批量删除多个标签

// ---------- 获取原始客户端（仅 Redis 驱动）----------
rdb := cache.GetRedis()
```

## 缓存标签

缓存标签功能参考 think-cache 设计，可将多个缓存项关联到标签下，方便批量管理。

### 标签操作

| 方法 | 功能 |
|------|------|
| `SetWithTags(key, value, tags, expiration)` | 设置缓存并关联多个标签 |
| `Tag(name)` | 获取标签操作器 |
| `Tag(name).Get()` | 获取标签下的所有缓存键 |
| `Tag(name).Flush()` | 删除标签下所有缓存（同时删除标签） |
| `Tag(name).Clear()` | 清除标签（不删除缓存） |
| `Tag(name).Has()` | 判断标签下是否有缓存键 |
| `Tag(name).Add(keys)` | 给标签添加缓存键 |
| `Tag(name).Remove(keys)` | 从标签中移除缓存键 |
| `FlushTags(tags...)` | 批量删除多个标签 |

### 使用示例

```go
// 设置缓存并关联标签
cache.SetWithTags("user:123", userData, []string{"user", "user:123"}, 3600)
cache.SetWithTags("user:456", userData2, []string{"user", "user:456"}, 3600)

// 更新用户信息后清除所有用户相关缓存
cache.Tag("user").Flush()

// 批量删除多个标签
cache.FlushTags("user", "article", "config")

// 给已有缓存添加标签
cache.Tag("hot").Add("user:123", "user:456")
```

## 设计要点

- 所有操作都会自动在 key 前面拼接 `prefix`，避免多项目冲突
- 支持三种驱动模式，可根据场景灵活选择
- 内存驱动适合开发测试，进程退出后数据丢失
- 文件驱动适合需要持久化但不想依赖 Redis 的场景
- Redis 驱动适合生产环境和分布式场景
- 获取原始 `*redis.Client` 后 **不会** 自动加前缀，操作 key 时请自行拼接
- Redis 初始化失败不会阻塞程序（只会 `logger.Warn`），方便本地调试跳过 Redis