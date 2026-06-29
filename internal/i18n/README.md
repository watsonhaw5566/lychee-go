# i18n · 国际化支持

提供应用级别的国际化支持，支持多语言翻译、自动语言检测和灵活的翻译管理。

## 特性

- ✅ 支持多语言翻译文件（YAML 格式）
- ✅ 自动语言检测（Header / Query / Cookie）
- ✅ 支持占位符替换
- ✅ 支持回退到默认语言
- ✅ Gin 中间件集成
- ✅ 动态添加翻译

## 配置

在 `config.yml` 中配置 i18n：

```yaml
i18n:
  default: zh-CN           # 默认语言
  dir: resources/lang      # 语言文件目录
```

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `default` | string | zh-CN | 默认语言代码 |
| `dir` | string | resources/lang | 语言文件存放目录 |

## 语言文件格式

在 `resources/lang/` 目录下创建语言文件，文件名格式为 `{lang}.yml`：

```yaml
# resources/lang/zh-CN.yml
hello: "你好"
welcome: "欢迎来到 Lychee-Go"
greeting: "你好，%s！"
error:
  not_found: "资源未找到"
  server_error: "服务器内部错误"
```

```yaml
# resources/lang/en.yml
hello: "Hello"
welcome: "Welcome to Lychee-Go"
greeting: "Hello, %s!"
error:
  not_found: "Resource not found"
  server_error: "Internal server error"
```

## API 使用

```go
import "lychee-go/internal/i18n"

// 获取翻译（使用当前语言）
msg := i18n.Get("hello")           // "你好"
msg := i18n.T("hello")             // 简写形式

// 带占位符的翻译
msg := i18n.Get("greeting", "Alice")  // "你好，Alice！"

// 指定语言获取翻译
msg := i18n.ForLang("en").Get("hello")  // "Hello"
msg := i18n.ForLang("en").T("greeting", "Bob")  // "Hello, Bob!"

// 获取当前语言
lang := i18n.GetLang()

// 设置当前语言
i18n.SetLang("en")

// 获取支持的语言列表
langs := i18n.GetAvailableLangs()
```

## Gin 中间件集成

```go
import (
    "lychee-go/internal/i18n"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()
    
    // 注册 i18n 中间件（自动检测语言）
    r.Use(i18n.Middleware())
    
    r.GET("/hello", func(c *gin.Context) {
        // 从上下文获取翻译
        msg := i18n.GetFromContext(c, "hello")
        c.JSON(200, gin.H{"message": msg})
    })
    
    r.Run()
}
```

## 语言检测优先级

中间件按以下顺序检测语言：

1. **Accept-Language Header**：`Accept-Language: zh-CN,zh;q=0.9,en;q=0.8`
2. **Query 参数**：`?lang=en`
3. **Cookie**：`lang=en`
4. **默认语言**：配置文件中的 `i18n.default`

## 完整示例

```go
package main

import (
    "lychee-go/internal/i18n"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()
    r.Use(i18n.Middleware())
    
    r.GET("/greet", func(c *gin.Context) {
        name := c.Query("name")
        if name == "" {
            name = "Guest"
        }
        
        // 使用上下文语言获取翻译
        msg := i18n.GetFromContext(c, "greeting", name)
        c.JSON(200, gin.H{"message": msg})
    })
    
    r.GET("/error", func(c *gin.Context) {
        errMsg := i18n.GetFromContext(c, "error.not_found")
        c.JSON(404, gin.H{"error": errMsg})
    })
    
    r.Run(":8080")
}
```

## 核心 API

| 方法 | 说明 | 参数 | 返回值 |
|------|------|------|--------|
| `Get(key, args...)` | 获取翻译（当前语言） | `key string, args ...interface{}` | `string` |
| `T(key, args...)` | Get 的简写形式 | `key string, args ...interface{}` | `string` |
| `ForLang(lang)` | 返回指定语言的翻译器 | `lang string` | `*I18n` |
| `SetLang(lang)` | 设置当前语言 | `lang string` | 无 |
| `GetLang()` | 获取当前语言 | 无 | `string` |
| `SetDefaultLang(lang)` | 设置默认语言 | `lang string` | 无 |
| `GetDefaultLang()` | 获取默认语言 | 无 | `string` |
| `GetAvailableLangs()` | 获取所有支持的语言 | 无 | `[]string` |
| `AddTranslation(lang, key, value)` | 动态添加翻译 | `lang, key, value string` | 无 |
| `GetFromContext(c, key, args...)` | 从 Gin 上下文获取翻译 | `c *gin.Context, key string, args ...interface{}` | `string` |
| `Middleware()` | Gin 中间件 | 无 | `gin.HandlerFunc` |

## 目录结构

```
resources/
└── lang/
    ├── zh-CN.yml    # 简体中文
    ├── zh-TW.yml    # 繁体中文
    ├── en.yml       # 英语
    ├── ja.yml       # 日语
    └── ko.yml       # 韩语
```

## 使用场景

1. **多语言网站**：根据用户语言偏好显示不同语言内容
2. **国际化 API**：根据请求头返回对应语言的错误消息
3. **后台管理**：支持管理员切换界面语言
4. **动态内容**：运行时动态添加翻译

## 注意事项

- 语言文件必须是 YAML 格式，键值对结构
- 支持嵌套键（如 `error.not_found`）
- 占位符使用 `%s`、`%d` 等标准 fmt 格式
- 未找到的翻译键会返回键名本身
- 推荐的语言代码格式：`zh-CN`、`en`、`ja` 等