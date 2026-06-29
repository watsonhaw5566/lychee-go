# cache · 缓存

基于 [go-redis](https://github.com/redis/go-redis) 封装的 Redis 缓存，带 key 前缀。

## 用法

```go
import "lychee-go/internal/cache"

// ---------- 基本 KV ----------
cache.Set("user:123", userData, 3600)           // 存，TTL 单位秒
data, err := cache.Get("user:123")              // 取
cache.Del("user:123")                           // 删
cache.Exists("user:123")                        // 是否存在

// ---------- 自增 ----------
cache.Incr("counter:views")                     // +1
cache.IncrBy("counter:views", 10)               // +10

// ---------- 过期 ----------
cache.Expire("session:abc", 7200)               // 设置过期时间（秒）
cache.TTL("session:abc")                        // 获取剩余 TTL

// ---------- 获取原始客户端 ----------
rdb := cache.GetRedis()                         // *redis.Client，可调用任意 go-redis API
rdb.HSet("profile", "name", "alice")
rdb.ZRange("leaderboard", 0, 10)
```

## 配置

```yaml
cache:
  driver: redis
  host: 127.0.0.1
  port: 6379
  password: ""
  database: 0
  prefix: lychee_go_      # 自动给每个 key 加前缀，避免多项目冲突
  default_ttl: 3600       # 默认 TTL（秒）
```

## 设计要点

- 所有 `Set/Get/Del/Exists/Incr` 等操作都会自动在 key 前面拼接 `prefix`。
- 获取原始 `*redis.Client` 后 **不会** 自动加前缀，操作 key 时请自行拼接。
- Redis 初始化失败不会阻塞程序（只会 `logger.Warn`），方便本地调试跳过 Redis。
