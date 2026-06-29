# config · 配置管理

基于 [viper](https://github.com/spf13/viper) 封装的配置读取器，支持单一 YAML 文件 + 环境变量覆盖。

## 用法

```go
import "lychee-go/internal/config"

// 在程序入口处初始化一次（通常由 boot.Boot() 调用）
config.Init("config/config.yml")

// 读取值（带默认参数）
port := config.GetInt("app.port", 8080)
name := config.GetString("app.name", "lychee-go")
debug := config.GetBool("app.debug", true)
ttl := config.GetInt("cache.default_ttl", 3600)
```

## 支持的读取方法

| 方法 | 说明 |
|------|------|
| `GetString(key, default)` | 字符串 |
| `GetInt(key, default)` | 整数 |
| `GetBool(key, default)` | 布尔 |
| `GetInt64(key, default)` | int64 |
| `GetFloat64(key, default)` | 浮点数 |
| `Get(key)` | 任意类型（返回 `interface{}`） |
| `GetStringSlice(key, sep)` | 分割字符串为数组 |

## 嵌套键

YAML 的嵌套结构用 `.` 分隔：

```yaml
database:
  host: 127.0.0.1
  port: 3306
```

读取：
```go
host := config.GetString("database.host", "127.0.0.1")
port := config.GetInt("database.port", 3306)
```

## 环境变量覆盖

设置 `LYCHEE_APP_PORT=9000` 可以覆盖 `config.yml` 中的 `app.port`（viper 的自动绑定）。
