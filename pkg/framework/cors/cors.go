package cors

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"lychee-go/pkg/framework/config"
	"lychee-go/pkg/framework/logger"

	"github.com/gin-gonic/gin"
)

// ======== 配置结构 ========

type Config struct {
	AllowOrigins     []string      // 允许的来源域名列表
	AllowMethods     []string      // 允许的 HTTP 方法
	AllowHeaders     []string      // 允许的请求头
	ExposeHeaders    []string      // 暴露给前端的响应头
	AllowCredentials bool          // 是否允许携带 Cookie
	MaxAge           time.Duration // Preflight 缓存时间
}

var (
	defaultConfig *Config
	once          sync.Once
)

// ======== 初始化 ========

func Init() *Config {
	once.Do(func() {
		defaultConfig = loadConfigFromYAML()
		logger.Info("[CORS] Initialized (origins: %s, methods: %s)",
			strings.Join(defaultConfig.AllowOrigins, ", "),
			strings.Join(defaultConfig.AllowMethods, ", "))
	})
	return defaultConfig
}

func loadConfigFromYAML() *Config {
	allowOrigins := parseList(config.GetString("cors.allow_origins", "*"))
	allowMethods := parseList(config.GetString("cors.allow_methods", "GET,POST,PUT,DELETE,OPTIONS"))
	allowHeaders := parseList(config.GetString("cors.allow_headers", "Content-Type,Authorization,X-Requested-With"))
	exposeHeaders := parseList(config.GetString("cors.expose_headers", ""))
	allowCredentials := config.GetBool("cors.allow_credentials", true)
	maxAgeSec := config.GetInt("cors.max_age", 86400)

	if len(allowOrigins) == 0 {
		allowOrigins = []string{"*"}
	}
	if len(allowMethods) == 0 {
		allowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}
	if len(allowHeaders) == 0 {
		allowHeaders = []string{"Content-Type", "Authorization", "X-Requested-With"}
	}

	return &Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     allowMethods,
		AllowHeaders:     allowHeaders,
		ExposeHeaders:    exposeHeaders,
		AllowCredentials: allowCredentials,
		MaxAge:           time.Duration(maxAgeSec) * time.Second,
	}
}

func parseList(input string) []string {
	if input == "" {
		return nil
	}
	parts := strings.Split(input, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// ======== 工具函数 ========

func (c *Config) isOriginAllowed(origin string) bool {
	if len(c.AllowOrigins) == 0 {
		return true
	}
	for _, allowed := range c.AllowOrigins {
		if allowed == "*" {
			return true
		}
		// 支持通配符，例如 *.example.com
		if strings.Contains(allowed, "*") {
			pattern := strings.ReplaceAll(allowed, ".", "\\.")
			pattern = strings.ReplaceAll(pattern, "*", ".*")
			if match, _ := matchPattern(pattern, origin); match {
				return true
			}
		}
		if strings.EqualFold(allowed, origin) {
			return true
		}
	}
	return false
}

func matchPattern(pattern, input string) (bool, error) {
	if !strings.HasPrefix(pattern, "^") {
		pattern = "^" + pattern
	}
	if !strings.HasSuffix(pattern, "$") {
		pattern = pattern + "$"
	}
	// 使用简单的字符串匹配（避免引入更多依赖）
	simple := strings.ReplaceAll(pattern, "\\.", ".")
	simple = strings.ReplaceAll(simple, ".*", "*")
	// 退化为后缀匹配
	suffix := strings.TrimPrefix(simple, "^")
	suffix = strings.TrimSuffix(suffix, "$")
	suffix = strings.TrimPrefix(suffix, "*")
	if strings.HasSuffix(input, suffix) {
		return true, nil
	}
	return strings.EqualFold(pattern, input), nil
}

// ======== Gin 中间件 ========

func Middleware(cfg ...*Config) gin.HandlerFunc {
	config := defaultConfig
	if len(cfg) > 0 && cfg[0] != nil {
		config = cfg[0]
	}
	if config == nil {
		config = Init()
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// ------ 设置 Vary 头（必须，否则 CDN/浏览器缓存会出问题） ------
		c.Writer.Header().Add("Vary", "Origin")
		c.Writer.Header().Add("Vary", "Access-Control-Request-Method")
		c.Writer.Header().Add("Vary", "Access-Control-Request-Headers")

		// ------ 非跨域请求（没有 Origin），直接放行 ------
		if origin == "" {
			c.Next()
			return
		}

		// ------ 校验 Origin ------
		if !config.isOriginAllowed(origin) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Origin not allowed: " + origin,
			})
			return
		}

		// ------ 设置 Allow-Origin ------
		// 注意：AllowCredentials = true 时不能用 *
		if config.AllowCredentials && len(config.AllowOrigins) > 0 && config.AllowOrigins[0] != "*" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else if len(config.AllowOrigins) == 1 && config.AllowOrigins[0] == "*" && !config.AllowCredentials {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		// ------ 设置其他 CORS 头 ------
		c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
		c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))

		if len(config.ExposeHeaders) > 0 {
			c.Writer.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
		}

		if config.AllowCredentials {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// ------ Preflight (OPTIONS) 请求 ------
		if c.Request.Method == "OPTIONS" {
			if config.MaxAge > 0 {
				c.Writer.Header().Set("Access-Control-Max-Age", itoa(int(config.MaxAge.Seconds())))
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// ------ 正常请求 ------
		c.Next()
	}
}

// AllowAll 一个方便的中间件：完全放行所有跨域请求（开发环境用）
func AllowAll() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Add("Vary", "Origin")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Restrict 限制特定 Origin 白名单
func Restrict(origins ...string) gin.HandlerFunc {
	cfg := &Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-Requested-With"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	return Middleware(cfg)
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
