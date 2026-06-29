# command · 可扩展命令行

一个让 `lychee-go` 支持 `lychee-go <command>` 子命令的小框架。

典型场景：

- `lychee-go migrate` —— 数据库迁移
- `lychee-go seed` —— 填充测试数据
- `lychee-go cron:run` —— 手动触发定时任务
- `lychee-go hello` —— 内置问候示例

---

## 1. 工作流程

```
命令行输入                      main.go                                  你的命令
┌──────────────────┐     ┌──────────────────────┐               ┌──────────────────┐
│ lychee-go hello  │ ──▶│ handleLightweight()  │ ─── no ──▶    │ boot.Boot()       │
└──────────────────┘     │ 拦截 help/version    │               │ 注册所有命令      │
                          │                       │               │ command.Dispatch │
                          │                       │               │   └─> hello.Run  │
                          └──────────────────────┘               └──────────────────┘
```

- **轻量级命令**（`help / version`）在 boot 之前直接运行，不会尝试连数据库 / Redis。
- **自定义命令**（`hello / migrate / ...`）会先执行完整的 `boot.Boot()`，所以你的 `Run` 函数里可以直接使用 `db.DB` / `cache.Get` / `queue.Dispatch` 等所有框架能力。

---

## 2. 注册命令

### 方式 1：简单命令（`RegisterSimple`）

最常用，不需要 flags：

```go
package mycommands

import (
    "fmt"
    "lychee-go/internal/command"
    "lychee-go/internal/db"
)

func init() {
    command.RegisterSimple("seed", "向 users 表插入测试数据",
        func(args []string) error {
            // 这里可以直接使用 db.DB / cache.Get 等
            user := map[string]interface{}{"username": "testuser", "email": "test@example.com"}
            if err := db.DB.Table("users").Create(&user).Error; err != nil {
                return fmt.Errorf("seed failed: %w", err)
            }
            fmt.Println("Inserted test user")
            return nil
        },
    )
}
```

然后在任意被 main 引用的 `import` 路径里引入这个包即可（例如在 `app/route/route.go` 顶部写 `_ "lychee-go/app/commands"`）。

### 方式 2：带 flags 的命令（`Register`）

支持 `flag` 包的完整能力：

```go
package mycommands

import (
    "flag"
    "fmt"
    "lychee-go/internal/command"
)

func init() {
    command.Register(&command.Command{
        Name:        "greet",
        Description: "问候命令，支持 -up 大写",
        Usage:       "[-up] name...",
        Flags: func(fs *flag.FlagSet) {
            // 在 fs 上定义 flags
            fs.Bool("up", false, "把问候语转大写")
            fs.Int("times", 1, "重复次数")
        },
        Run: func(args []string) error {
            // args 是 flags 解析后剩余的位置参数
            // 想访问 flag 值，需要你在 Flags 函数外保存一个闭包变量
            return nil
        },
    })
}
```

> 小贴士：如果需要在 `Run` 里访问 `-up` 的值，可以用闭包：
>
> ```go
> up := false
> cmd := &command.Command{
>     Flags: func(fs *flag.FlagSet) { fs.BoolVar(&up, "up", false, "大写") },
>     Run:   func(args []string) error { fmt.Println("up =", up); return nil },
> }
> command.Register(cmd)
> ```

---

## 3. 内置命令

| 命令 | 说明 | 是否需要 boot |
|------|------|--------------|
| `help`, `-h`, `--help` | 显示帮助 | ❌ |
| `version`, `-v`, `--version` | 显示版本 | ❌ |
| `list` / `ls` / `commands` | 列出所有已注册命令 | ✅ |
| `hello` | 问候示例（`boot.InitCommands` 注册） | ✅ |
| `echo` | 回显参数（`boot.InitCommands` 注册） | ✅ |
| `doc` | 打印已注册 Swagger 端点 | ✅ |

运行示例：

```bash
lychee-go help              # 显示帮助
lychee-go list              # 列出所有命令
lychee-go hello Alice       # 打印 Hello, Alice!
lychee-go doc               # 打印已注册的 Swagger 端点
```

---

## 4. 关键设计点

1. **幂等注册**：`Register` 内部通过 map 去重，同名字符串重复注册不会报错（后注册的覆盖先注册的）。
2. **错误即退出**：`Command.Run` 返回 `error` 不为空时，`main.go` 会以 `log.Fatalf` 退出，并打印错误信息；所以你的 `Run` 里直接 `return err` 就好。
3. **完整运行环境**：自定义命令的 `Run` 被触发时，`boot.Boot()` 已完成，你可以直接使用 `db.DB` / `cache.Get` / `queue.Dispatch` / `cron.AddFunc` 等所有框架模块。
4. **零代码扩展**：把命令写在独立的包（如 `app/commands/`），然后在 `app/route/route.go` 或 `main.go` 中用 `_ import` 引入，init() 就会自动注册——不需要改任何启动代码。

---

## 5. 建议目录结构

```
lychee-go/
├─ main.go                    # 入口：command.Dispatch()
├─ internal/boot/boot.go      # 在这里注册内置的 hello / echo / doc
├─ internal/command/          # 框架本身（README 所在目录）
│
└─ app/commands/              # 你的业务命令（推荐）
   ├─ seed.go                 #     command.RegisterSimple("seed", ...)
   └─ migrate.go              #     command.RegisterSimple("migrate", ...)
```

在 `app/route/route.go` 顶部加一行：

```go
import _ "lychee-go/app/commands"
```

这样执行 `lychee-go seed` 时会自动发现该命令。
