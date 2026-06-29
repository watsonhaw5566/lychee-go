package boot

import (
	"flag"
	"fmt"
	"strings"

	"lychee-go/internal/cache"
	"lychee-go/internal/command"
	"lychee-go/internal/config"
	"lychee-go/internal/cookie"
	"lychee-go/internal/cors"
	"lychee-go/internal/cron"
	"lychee-go/internal/db"
	"lychee-go/internal/env"
	"lychee-go/internal/filesystem"
	"lychee-go/internal/i18n"
	"lychee-go/internal/jwt"
	"lychee-go/internal/logger"
	"lychee-go/internal/queue"
	"lychee-go/internal/session"
	"lychee-go/internal/swagger"
	"lychee-go/internal/throttle"
	"lychee-go/internal/websocket"
)

// ============================================================
// 模块初始化决策表：判断哪些模块需要初始化
// ============================================================

// shouldInitDatabase 判断是否配置了数据库
func shouldInitDatabase() bool {
	// 必须配置了 database 块，且 host / database 非空
	return config.IsConfigured("database.host") &&
		config.IsConfigured("database.database")
}

// shouldInitCache 判断是否配置了缓存
func shouldInitCache() bool {
	// 必须配置了 cache 块，且 host 非空，driver 不为 none/disabled
	if !config.IsConfigured("cache.driver") || !config.IsConfigured("cache.host") {
		return false
	}
	driver := strings.ToLower(config.GetString("cache.driver", ""))
	return driver != "none" && driver != "disabled" && driver != ""
}

// shouldInitJWT 判断是否配置了 JWT
func shouldInitJWT() bool {
	// 必须配置了 auth.jwt_secret 非空
	return config.IsConfigured("auth.jwt_secret")
}

// shouldInitFilesystem 判断是否配置了文件系统
func shouldInitFilesystem() bool {
	// 必须配置了 filesystem.default 非空，且对应驱动配置存在
	if !config.IsConfigured("filesystem.default") {
		return false
	}
	defaultDriver := config.GetString("filesystem.default", "")
	return defaultDriver != "" && defaultDriver != "none" && defaultDriver != "disabled"
}

// shouldInitQueue 判断是否配置了消息队列
func shouldInitQueue() bool {
	if !config.IsConfigured("queue.driver") {
		return false
	}
	driver := strings.ToLower(config.GetString("queue.driver", ""))
	return driver != "none" && driver != "disabled" && driver != ""
}

// shouldInitCron 判断是否配置了定时任务
func shouldInitCron() bool {
	// cron 是纯内存调度器，默认启用；只有显式设置 enabled: false 才禁用
	if config.IsSet("cron.enabled") {
		return config.GetBool("cron.enabled", true)
	}
	return true
}

// shouldInitSession 判断是否配置了 Session
func shouldInitSession() bool {
	// session 是纯内存 KV，默认启用；只有显式设置 ttl = 0 或 enabled = false 才禁用
	if config.IsSet("session.enabled") {
		return config.GetBool("session.enabled", true)
	}
	return config.GetInt("session.ttl", 0) > 0
}

// shouldInitCORS 判断是否配置了 CORS
func shouldInitCORS() bool {
	// 必须配置了 cors 块且 allow_origins 非空，且不为 none/disabled
	if !config.IsConfigured("cors.allow_origins") {
		return false
	}
	origins := strings.ToLower(config.GetString("cors.allow_origins", ""))
	return origins != "none" && origins != "disabled" && origins != ""
}

// shouldInitThrottle 判断是否配置了限流
func shouldInitThrottle() bool {
	// 必须配置了 throttle.limit 且大于 0
	if config.IsSet("throttle.enabled") {
		return config.GetBool("throttle.enabled", true)
	}
	return config.GetInt("throttle.limit", 0) > 0 &&
		config.GetInt("throttle.window", 0) > 0
}

// shouldInitCookie 判断是否配置了 Cookie
func shouldInitCookie() bool {
	// 必须配置了 cookie 块（只要 cookie.secret 有配置就算）
	return config.IsConfigured("cookie.secret")
}

// shouldInitSwagger 判断是否启用 Swagger
func shouldInitSwagger() bool {
	// 默认启用；只有显式 swagger.enabled: false 才禁用
	if config.IsSet("swagger.enabled") {
		return config.GetBool("swagger.enabled", true)
	}
	return true
}

// shouldInitWebSocket 判断是否启用 WebSocket
func shouldInitWebSocket() bool {
	// 默认启用；只有显式 websocket.enabled: false 才禁用
	if config.IsSet("websocket.enabled") {
		return config.GetBool("websocket.enabled", true)
	}
	return true
}

// shouldInitI18n 判断是否启用国际化
func shouldInitI18n() bool {
	// 默认启用；只有显式 i18n.enabled: false 才禁用
	if config.IsSet("i18n.enabled") {
		return config.GetBool("i18n.enabled", true)
	}
	return true
}

// shouldInitCommand 判断是否启用命令行
// 命令行模块由 main.go 通过 Dispatch() 直接调用，不通过 Boot；
// 但此函数用于 boot 启动日志里显示"已注册 N 个命令"
func shouldInitCommand() bool {
	return true
}

// InitCommands 注册内置示例命令（可以在这里加入更多）
func InitCommands() {
	// 示例 1：hello
	command.RegisterSimple("hello", "问候示例：打印 Hello", func(args []string) error {
		name := "World"
		if len(args) > 0 {
			name = args[0]
		}
		fmt.Printf("Hello, %s! 👋\n", name)
		return nil
	})

	// 示例 2：echo
	command.Register(&command.Command{
		Name:        "echo",
		Description: "回显参数（带 -up 大写参数）",
		Usage:       "[-up] text...",
		Flags: func(fs *flag.FlagSet) {
			fs.Bool("up", false, "转大写")
		},
		Run: func(args []string) error {
			fmt.Println(strings.Join(args, " "))
			return nil
		},
	})

	// 示例 3：doc —— 打印已注册的 Swagger 端点
	command.RegisterSimple("doc", "打印已注册的 Swagger API 端点", func(args []string) error {
		swagger.Init()
		swagger.PrintSummary()
		return nil
	})
}

// ============================================================
// Boot 框架完整启动引导
// ============================================================
func Boot(configPath string) error {
	// 0. 加载环境变量（必须最先，因为 config.yml 可能引用环境变量）
	env.Load()

	// 1. 加载配置
	if err := config.Init(configPath); err != nil {
		return err
	}

	// 2. 初始化日志（基础模块，无条件）
	logger.Init()
	logger.Info("============================================================")
	logger.Info("Lychee-Go Framework starting...")
	logger.Info("============================================================")
	logger.Info("App Name: %s, Version: %s",
		config.GetString("app.name", "lychee-go"),
		config.GetString("app.version", "1.0.0"))
	logger.Info("Debug Mode: %v, Port: %d",
		config.GetBool("app.debug", true),
		config.GetInt("app.port", 8080))

	// 存储启动结果，用于最终打印
	results := []string{}
	addResult := func(name, status, detail string) {
		if detail != "" {
			results = append(results, fmt.Sprintf("  - %-12s %s (%s)", name, status, detail))
		} else {
			results = append(results, fmt.Sprintf("  - %-12s %s", name, status))
		}
	}

	// 3. 初始化数据库
	if shouldInitDatabase() {
		if err := db.Init(); err != nil {
			logger.Warn("Database init failed: %v (DB features will not work)", err)
			addResult("Database", "FAILED", err.Error())
		} else {
			addResult("Database", "OK",
				fmt.Sprintf("%s@%s:%d",
					config.GetString("database.driver", "mysql"),
					config.GetString("database.host", ""),
					config.GetInt("database.port", 3306)))
		}
	} else {
		logger.Info("Database skipped: not configured")
		addResult("Database", "SKIPPED", "not configured")
	}

	// 4. 初始化缓存
	if shouldInitCache() {
		if err := cache.Init(); err != nil {
			logger.Warn("Cache init failed: %v (Redis features will not work)", err)
			addResult("Cache", "FAILED", err.Error())
		} else {
			addResult("Cache", "OK",
				fmt.Sprintf("%s@%s:%d",
					config.GetString("cache.driver", "redis"),
					config.GetString("cache.host", ""),
					config.GetInt("cache.port", 6379)))
		}
	} else {
		logger.Info("Cache skipped: not configured")
		addResult("Cache", "SKIPPED", "not configured")
	}

	// 5. 初始化文件系统
	if shouldInitFilesystem() {
		if err := filesystem.Init(); err != nil {
			logger.Warn("Filesystem init failed: %v", err)
			addResult("Filesystem", "FAILED", err.Error())
		} else {
			addResult("Filesystem", "OK",
				fmt.Sprintf("driver: %s", config.GetString("filesystem.default", "local")))
		}
	} else {
		logger.Info("Filesystem skipped: not configured")
		addResult("Filesystem", "SKIPPED", "not configured")
	}

	// 6. 初始化 JWT 鉴权
	if shouldInitJWT() {
		jwt.Init()
		addResult("JWT", "OK",
			fmt.Sprintf("ttl: %ds, max_per_user: %d",
				config.GetInt("auth.jwt_ttl", 86400),
				config.GetInt("auth.max_per_user", 10)))
	} else {
		logger.Info("JWT skipped: jwt_secret not configured")
		addResult("JWT", "SKIPPED", "jwt_secret not set")
	}

	// 7. 初始化 Session
	if shouldInitSession() {
		session.Init()
		addResult("Session", "OK",
			fmt.Sprintf("ttl: %ds", config.GetInt("session.ttl", 7200)))
	} else {
		logger.Info("Session skipped: ttl=0 or disabled")
		addResult("Session", "SKIPPED", "ttl=0 or disabled")
	}

	// 8. 初始化消息队列
	if shouldInitQueue() {
		// queue 需要 redis client，如果 cache 未启用则传 nil
		// 注意：如果 queue.driver = redis 但 cache 未启用，这里会用 nil 尝试；
		// queue 内部会根据 driver 选择内存实现，所以安全
		rdb := cache.GetRedis() // nil safe，如果 cache 未初始化则返回 nil
		if err := queue.Init(rdb); err != nil {
			logger.Warn("Queue init failed: %v", err)
			addResult("Queue", "FAILED", err.Error())
		} else {
			addResult("Queue", "OK",
				fmt.Sprintf("driver: %s", config.GetString("queue.driver", "memory")))
		}
	} else {
		logger.Info("Queue skipped: not configured")
		addResult("Queue", "SKIPPED", "not configured")
	}

	// 9. 启动定时任务调度器
	if shouldInitCron() {
		cron.Start()
		addResult("Cron", "OK",
			fmt.Sprintf("%d tasks registered", cron.TaskCount()))
	} else {
		logger.Info("Cron skipped: disabled by config")
		addResult("Cron", "SKIPPED", "disabled")
	}

	// 10. 初始化 CORS
	if shouldInitCORS() {
		cors.Init()
		addResult("CORS", "OK", "")
	} else {
		logger.Info("CORS skipped: allow_origins not configured")
		addResult("CORS", "SKIPPED", "not configured")
	}

	// 11. 初始化限流 Throttle
	if shouldInitThrottle() {
		throttle.Init(nil) // 默认内存存储
		addResult("Throttle", "OK",
			fmt.Sprintf("limit: %d/%ds",
				config.GetInt("throttle.limit", 60),
				config.GetInt("throttle.window", 60)))
	} else {
		logger.Info("Throttle skipped: not configured")
		addResult("Throttle", "SKIPPED", "limit=0 or disabled")
	}

	// 12. 初始化 Cookie
	if shouldInitCookie() {
		cookie.Init()
		addResult("Cookie", "OK",
			fmt.Sprintf("httponly: %v, samesite: %s",
				config.GetBool("cookie.httponly", true),
				config.GetString("cookie.samesite", "lax")))
	} else {
		logger.Info("Cookie skipped: secret not configured")
		addResult("Cookie", "SKIPPED", "secret not set")
	}

	// 13. 初始化 Swagger（生成 API 文档）
	if shouldInitSwagger() {
		swagger.Init()
		addResult("Swagger", "OK",
			fmt.Sprintf("path: %s", config.GetString("swagger.path", "/swagger")))
	} else {
		logger.Info("Swagger skipped: disabled by config")
		addResult("Swagger", "SKIPPED", "disabled")
	}

	// 14. 初始化 WebSocket
	if shouldInitWebSocket() {
		websocket.Init()
		addResult("WebSocket", "OK",
			fmt.Sprintf("read_buffer: %d, write_buffer: %d",
				config.GetInt("websocket.read_buffer_size", 1024),
				config.GetInt("websocket.write_buffer_size", 1024)))
	} else {
		logger.Info("WebSocket skipped: disabled by config")
		addResult("WebSocket", "SKIPPED", "disabled")
	}

	// 15. 初始化 i18n 国际化
	if shouldInitI18n() {
		i18n.Init()
		addResult("I18n", "OK",
			fmt.Sprintf("default: %s", config.GetString("i18n.default", "zh-CN")))
	} else {
		logger.Info("I18n skipped: disabled by config")
		addResult("I18n", "SKIPPED", "disabled")
	}

	// 16. 注册内置 Command 命令
	InitCommands()
	addResult("Commands", "OK",
		fmt.Sprintf("%d commands registered", len(command.List())))

	// ======== 完成 ========
	logger.Info("============================================================")
	logger.Info("Lychee-Go Framework boot completed!")
	logger.Info("Logger:     OK (always-on)")
	for _, line := range results {
		logger.Info(line)
	}
	logger.Info("============================================================")

	return nil
}
