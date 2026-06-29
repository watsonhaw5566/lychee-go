package throttle

import (
	"sync"
	"sync/atomic"
	"time"

	"lychee-go/pkg/framework/config"
	"lychee-go/pkg/framework/logger"

	"github.com/gin-gonic/gin"
)

// ======== 限流存储接口 ========

type Store interface {
	Take(key string, limit int, window time.Duration) (bool, int, time.Time)
	Reset(key string)
	Stats(key string, limit int, window time.Duration) (int, time.Time)
}

// ======== 内存存储（令牌桶算法 + 滑动窗口） ========

type memoryBucket struct {
	tokens     int64
	lastRefill int64
	expiresAt  time.Time
}

type MemoryStore struct {
	buckets sync.Map
}

func NewMemoryStore() *MemoryStore {
	store := &MemoryStore{}
	// 后台 goroutine 清理过期桶
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			store.buckets.Range(func(key, value interface{}) bool {
				b := value.(*memoryBucket)
				if now.After(b.expiresAt) {
					store.buckets.Delete(key)
				}
				return true
			})
		}
	}()
	return store
}

func (s *MemoryStore) Take(key string, limit int, window time.Duration) (bool, int, time.Time) {
	nowNano := time.Now().UnixNano()
	windowNano := window.Nanoseconds()
	resetAt := time.Unix(0, nowNano).Add(window)

	val, _ := s.buckets.LoadOrStore(key, &memoryBucket{
		tokens:     int64(limit) - 1,
		lastRefill: nowNano,
		expiresAt:  resetAt,
	})
	bucket := val.(*memoryBucket)

	for {
		current := atomic.LoadInt64(&bucket.tokens)
		last := atomic.LoadInt64(&bucket.lastRefill)

		elapsed := float64(nowNano-last) / float64(windowNano)
		refill := int64(float64(limit) * elapsed)

		var newTokens int64
		if refill > 0 {
			newTokens = current + refill
			if newTokens > int64(limit) {
				newTokens = int64(limit)
			}
			atomic.StoreInt64(&bucket.lastRefill, nowNano)
		} else {
			newTokens = current
		}

		if newTokens <= 0 {
			return false, 0, time.Unix(0, last).Add(window)
		}

		if atomic.CompareAndSwapInt64(&bucket.tokens, current, newTokens-1) {
			bucket.expiresAt = time.Now().Add(window)
			return true, int(newTokens - 1), resetAt
		}
	}
}

func (s *MemoryStore) Reset(key string) {
	s.buckets.Delete(key)
}

func (s *MemoryStore) Stats(key string, limit int, window time.Duration) (int, time.Time) {
	val, ok := s.buckets.Load(key)
	if !ok {
		return limit, time.Now().Add(window)
	}
	bucket := val.(*memoryBucket)
	remaining := int(atomic.LoadInt64(&bucket.tokens))
	if remaining < 0 {
		remaining = 0
	}
	return remaining, time.Unix(0, bucket.lastRefill).Add(window)
}

// ======== 限流配置 ========

type Config struct {
	Limit       int
	Window      time.Duration
	KeyPrefix   string
	Message     string
	UseClientIP bool
}

var (
	defaultStore  Store
	defaultConfig *Config
	once          sync.Once
)

// ======== 初始化 ========

func Init(_ ...interface{}) {
	once.Do(func() {
		limit := config.GetInt("throttle.limit", 60)
		windowSec := config.GetInt("throttle.window", 60)
		prefix := config.GetString("throttle.prefix", "lychee_go_")

		defaultConfig = &Config{
			Limit:       limit,
			Window:      time.Duration(windowSec) * time.Second,
			KeyPrefix:   prefix,
			Message:     "请求过于频繁，请稍后再试",
			UseClientIP: true,
		}

		defaultStore = NewMemoryStore()
		logger.Info("[Throttle] Initialized (Memory, limit: %d/%ds)", limit, windowSec)
	})
}

func SetStore(store Store) {
	defaultStore = store
}

// ======== 便捷的预配置 ========

func PerMinute(n int) *Config {
	return &Config{Limit: n, Window: time.Minute, Message: "请求过于频繁，请稍后再试", UseClientIP: true}
}

func PerSecond(n int) *Config {
	return &Config{Limit: n, Window: time.Second, Message: "请求过于频繁，请稍后再试", UseClientIP: true}
}

func PerHour(n int) *Config {
	return &Config{Limit: n, Window: time.Hour, Message: "请求过于频繁，请稍后再试", UseClientIP: true}
}

func Login() *Config {
	return &Config{Limit: 10, Window: 5 * time.Minute, KeyPrefix: "login:", Message: "登录尝试次数过多，请 5 分钟后再试", UseClientIP: true}
}

func API() *Config {
	return &Config{Limit: 60, Window: time.Minute, KeyPrefix: "api:", Message: "API 请求过于频繁，请稍后再试", UseClientIP: true}
}

// ======== Gin 中间件 ========

func Middleware(cfg ...*Config) gin.HandlerFunc {
	c := defaultConfig
	if len(cfg) > 0 && cfg[0] != nil {
		c = cfg[0]
	}
	if c == nil {
		c = &Config{Limit: 60, Window: time.Minute, Message: "请求过于频繁，请稍后再试", UseClientIP: true}
	}

	store := defaultStore
	if store == nil {
		store = NewMemoryStore()
	}

	return func(ctx *gin.Context) {
		key := buildKey(ctx, c)
		allowed, remaining, resetAt := store.Take(key, c.Limit, c.Window)

		ctx.Header("X-RateLimit-Limit", itoa(c.Limit))
		ctx.Header("X-RateLimit-Remaining", itoa(maxInt(0, remaining)))
		ctx.Header("X-RateLimit-Reset", itoa64(resetAt.Unix()))

		if !allowed {
			retryAfter := int(time.Until(resetAt).Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}
			ctx.Header("Retry-After", itoa(retryAfter))
			ctx.AbortWithStatusJSON(429, gin.H{
				"code":                429,
				"message":             c.Message,
				"retry_after_seconds": retryAfter,
				"reset_at":            resetAt.Format(time.RFC3339),
			})
			return
		}

		ctx.Next()
	}
}

func buildKey(c *gin.Context, cfg *Config) string {
	var parts []string
	if cfg.KeyPrefix != "" {
		parts = append(parts, cfg.KeyPrefix)
	}

	if cfg.UseClientIP {
		parts = append(parts, clientIP(c))
	} else {
		parts = append(parts, c.Request.Method, c.FullPath())
	}

	if userID := c.GetHeader("X-User-ID"); userID != "" {
		parts = append(parts, "user:"+userID)
	}

	result := ""
	for i, p := range parts {
		if i == 0 {
			result = p
		} else {
			result += "_" + p
		}
	}
	return result
}

func clientIP(c *gin.Context) string {
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}
	ip := c.ClientIP()
	if ip == "" || ip == "::1" {
		return "127.0.0.1"
	}
	return ip
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func itoa64(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ======== 便捷组合中间件 ========

func Global(limit int, window time.Duration) gin.HandlerFunc {
	return Middleware(&Config{Limit: limit, Window: window, KeyPrefix: "global:", Message: "服务器繁忙，请稍后再试", UseClientIP: false})
}

func PerRoute(limit int, window time.Duration) gin.HandlerFunc {
	return Middleware(&Config{Limit: limit, Window: window, KeyPrefix: "route:", Message: "该接口请求过于频繁", UseClientIP: true})
}

func PerUser(limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		userKey := ""
		if id := c.GetHeader("X-User-ID"); id != "" {
			userKey = id
		} else if v, ok := c.Get("user_id"); ok {
			userKey = itoa64(int64(v.(uint)))
		}
		if userKey == "" {
			userKey = clientIP(c)
		}

		store := defaultStore
		if store == nil {
			store = NewMemoryStore()
		}

		key := "user:" + userKey + ":" + c.FullPath()
		allowed, remaining, resetAt := store.Take(key, limit, window)

		c.Header("X-RateLimit-Limit", itoa(limit))
		c.Header("X-RateLimit-Remaining", itoa(maxInt(0, remaining)))
		c.Header("X-RateLimit-Reset", itoa64(resetAt.Unix()))

		if !allowed {
			retryAfter := int(time.Until(resetAt).Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.Header("Retry-After", itoa(retryAfter))
			c.AbortWithStatusJSON(429, gin.H{
				"code":                429,
				"message":             "该接口请求过于频繁",
				"retry_after_seconds": retryAfter,
			})
			return
		}

		c.Next()
	}
}

// ======== 手动控制 ========

func Allow(key string, limit int, window time.Duration) (bool, int, time.Time) {
	store := defaultStore
	if store == nil {
		store = NewMemoryStore()
	}
	return store.Take(key, limit, window)
}

func Reset(key string) {
	if defaultStore != nil {
		defaultStore.Reset(key)
	}
}
