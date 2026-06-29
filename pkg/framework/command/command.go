package command

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/watsonhaw5566/lychee-go/pkg/framework/logger"
)

// ============================================================
// Command 定义
// ============================================================

// Command 自定义命令
type Command struct {
	Name        string                    // 命令名（如 "migrate", "seed"）
	Description string                    // 简介
	Usage       string                    // 用法说明
	Flags       func(fs *flag.FlagSet)    // 定义 flags
	Run         func(args []string) error // 执行逻辑（参数 args 是去除 flags 后的剩余）
}

// ============================================================
// 注册表
// ============================================================

var (
	commands = make(map[string]*Command)
	mu       sync.RWMutex
)

// Register 注册一个自定义命令
//
//	swagger.Register(&command.Command{Name: "hello", ...})
func Register(cmd *Command) {
	if cmd == nil || cmd.Name == "" {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	commands[cmd.Name] = cmd
}

// Get 查找一个命令
func Get(name string) (*Command, bool) {
	mu.RLock()
	defer mu.RUnlock()
	cmd, ok := commands[name]
	return cmd, ok
}

// List 返回所有已注册的命令名（字母序）
func List() []string {
	mu.RLock()
	defer mu.RUnlock()
	names := make([]string, 0, len(commands))
	for n := range commands {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// ============================================================
// Dispatch: 解析命令行参数并执行对应命令
// ============================================================

// Dispatch 解析 os.Args，若存在可识别的子命令则执行之
//
// 返回值：
//   - 执行了命令 → (true, error)，调用者应直接退出程序
//   - 没有识别到命令 → (false, nil)，调用者继续进入 HTTP 服务器
//
// 用法（main.go）：
//
//	if ran, err := command.Dispatch(); ran {
//	    if err != nil { log.Fatal(err) }
//	    return
//	}
//	// 进入 HTTP 服务
func Dispatch() (ran bool, err error) {
	// 参数必须至少 1 个（子命令名）
	if len(os.Args) < 2 {
		return false, nil
	}

	// 第一个非 flag 参数当作命令名
	cmdName := os.Args[1]

	// 处理内置命令
	switch cmdName {
	case "-h", "--help", "help", "?", "":
		printHelp()
		return true, nil
	case "-v", "--version", "version":
		fmt.Println("lychee-go version 1.0.0")
		return true, nil
	case "list", "ls", "commands":
		printCommandList()
		return true, nil
	}

	// 查找用户自定义命令
	cmd, ok := Get(cmdName)
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmdName)
		printHelp()
		return true, fmt.Errorf("unknown command: %s", cmdName)
	}

	// 创建 flag set 解析该命令自己的参数
	fs := flag.NewFlagSet(cmdName, flag.ExitOnError)
	if cmd.Flags != nil {
		cmd.Flags(fs)
	}
	if err := fs.Parse(os.Args[2:]); err != nil {
		return true, err
	}

	logger.Info("[Command] Running: %s", cmdName)

	if cmd.Run == nil {
		fmt.Printf("Command '%s' has no Run function.\n", cmdName)
		return true, nil
	}

	if err := cmd.Run(fs.Args()); err != nil {
		logger.Error("[Command] Failed: %s: %v", cmdName, err)
		return true, err
	}

	logger.Info("[Command] Completed: %s", cmdName)
	return true, nil
}

// ============================================================
// 帮助信息
// ============================================================

func printHelp() {
	fmt.Println("========================================================")
	fmt.Println(" Lychee-Go CLI")
	fmt.Println("========================================================")
	fmt.Println()
	fmt.Println(" Usage:")
	fmt.Println("   lychee-go [command] [flags]")
	fmt.Println()
	fmt.Println(" 不带参数直接运行 → 启动 HTTP 服务器")
	fmt.Println()
	fmt.Println(" Built-in commands:")
	fmt.Println("   help, -h, --help      显示帮助")
	fmt.Println("   version, -v, --version 显示版本")
	fmt.Println("   list, ls, commands    列出所有已注册的命令")
	fmt.Println()
	names := List()
	if len(names) > 0 {
		fmt.Println(" Registered commands:")
		for _, n := range names {
			cmd, _ := Get(n)
			if cmd.Description != "" {
				fmt.Printf("   %-20s %s\n", n, cmd.Description)
			} else {
				fmt.Printf("   %s\n", n)
			}
		}
		fmt.Println()
		fmt.Println(" 查看命令的具体参数：")
		fmt.Println("   lychee-go <command> -h")
		fmt.Println()
	}
	fmt.Println(" Examples:")
	fmt.Println("   lychee-go                     # 启动 HTTP 服务器")
	fmt.Println("   lychee-go hello -name=Alice   # 运行 hello 命令")
	fmt.Println("   lychee-go list                # 列出所有命令")
	fmt.Println()
}

func printCommandList() {
	names := List()
	if len(names) == 0 {
		fmt.Println("No commands registered.")
		return
	}
	fmt.Printf("%d command(s) registered:\n", len(names))
	for _, n := range names {
		cmd, _ := Get(n)
		desc := cmd.Description
		if desc == "" {
			desc = "(no description)"
		}
		fmt.Printf("  %-20s %s\n", n, desc)
	}
}

// ============================================================
// 便捷 API：在业务代码中注册命令
// ============================================================

// RegisterSimple 注册一个简单命令（只需要名字、描述和 Run）
func RegisterSimple(name, description string, run func(args []string) error) {
	Register(&Command{
		Name:        name,
		Description: description,
		Run:         run,
	})
}

// PrintUsage 打印某个命令的用法（Command.Run 内部调用）
func PrintUsage(cmd *Command) {
	if cmd.Usage != "" {
		fmt.Println("Usage:")
		for _, line := range strings.Split(cmd.Usage, "\n") {
			fmt.Printf("  lychee-go %s %s\n", cmd.Name, line)
		}
	} else {
		fmt.Printf("Usage: lychee-go %s [flags]\n", cmd.Name)
	}
}
