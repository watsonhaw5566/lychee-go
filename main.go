package main

import (
	"fmt"
	"log"
	"os"

	"lychee-go/app/route"
	"lychee-go/internal/boot"
	"lychee-go/internal/command"
	"lychee-go/internal/config"
	"lychee-go/internal/swagger"

	"github.com/gin-gonic/gin"
)

func main() {
	// 0. 处理无需 boot 的命令（help / version / list 等内置命令）
	//    如果传入的参数是这些，直接运行然后退出，避免去尝试连 DB/Redis
	if handleLightweightCommands() {
		return
	}

	// 1. 框架启动引导（boot.Boot 内部会注册完整的命令集）
	if err := boot.Boot("./config"); err != nil {
		log.Fatalf("[lychee-go] Boot failed: %v", err)
	}

	// 2. 检查是否有需要 boot 之后才能运行的自定义命令
	//    例如 `lychee-go hello` —— command.Dispatch 会识别并运行
	if ran, err := command.Dispatch(); ran {
		if err != nil {
			log.Fatalf("[lychee-go] Command failed: %v", err)
		}
		return
	}

	// 3. 设置 Gin 运行模式
	if config.GetBool("app.debug", true) {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 4. 创建 Gin 引擎
	r := gin.New()

	// 5. 注册业务路由
	route.Register(r)

	// 6. 注册 Swagger 路由（/swagger 和 /swagger/doc.json）
	if config.GetBool("swagger.enabled", true) {
		swagger.RegisterRoutes(r)
		// 可选：自动扫描已注册的 Gin 路由，补充基础 Swagger 定义
		swagger.ScanGinRoutes(r)
		// 终端打印已注册端点摘要
		swagger.PrintSummary()
	}

	// 7. 启动服务
	addr := fmt.Sprintf(":%d", config.GetInt("app.port", 8080))
	log.Printf("[lychee-go] Server starting on http://localhost%s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("[lychee-go] Failed to start server: %v", err)
	}
}

// handleLightweightCommands 处理不需要完整 boot 的命令：help / version / list
// 返回 true 表示"已处理，程序应退出"；false 表示继续正常启动。
func handleLightweightCommands() bool {
	if len(os.Args) < 2 {
		return false
	}
	cmd := os.Args[1]
	switch cmd {
	case "-h", "--help", "help", "?":
		fmt.Println("========================================================")
		fmt.Println(" Lychee-Go CLI")
		fmt.Println("========================================================")
		fmt.Println()
		fmt.Println(" 不带参数 → 启动 HTTP 服务器")
		fmt.Println("   lychee-go help            显示帮助")
		fmt.Println("   lychee-go version         显示版本")
		fmt.Println("   lychee-go list            列出所有命令")
		fmt.Println("   lychee-go <command>       运行自定义命令")
		fmt.Println()
		fmt.Println("提示：运行自定义命令之前会先执行 boot 引导，所以会连 DB/Redis。")
		return true
	case "-v", "--version", "version":
		fmt.Println("lychee-go 1.0.0")
		return true
	case "list", "ls", "commands":
		// list 需要 boot 之后才能列出用户注册的命令，所以这里返回 false，
		// 由 command.Dispatch() 在 boot 之后处理
		return false
	}
	return false
}
