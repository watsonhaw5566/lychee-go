# env · 环境变量管理

用于解析 `.env` 配置文件，并支持在 `config.yml` 中通过 `${ENV_VAR}` 语法引用环境变量。

## 特性

- ✅ 支持 `.env` 文件解析
- ✅ 支持环境变量引用 `${VAR}` 和默认值 `${VAR:-default}`
- ✅ 支持在 config.yml 中使用环境变量
- ✅ 自动屏蔽敏感信息日志输出

## .env 文件格式

```env
# 基本格式
APP_NAME=lychee-go
APP_DEBUG=true
APP_PORT=8080

# 带引号的值
DATABASE_NAME="lychee_go"
API_KEY='secret-key-123'

# 环境变量引用（支持默认值）
DB_PASSWORD=${DB_PASSWORD:-default123}
OSS_ACCESS_KEY=${OSS_ACCESS_KEY:-}

# 注释
# 这是一行注释，不会被解析
```

## 在 config.yml 中使用

```yaml
app:
  name: ${APP_NAME:-lychee-go}
  debug: ${APP_DEBUG:-false}
  port: ${APP_PORT:-8080}

database:
  host: ${DB_HOST:-127.0.0.1}
  port: ${DB_PORT:-3306}
  username: ${DB_USERNAME:-root}
  password: ${DB_PASSWORD:-}

filesystem:
  oss:
    access_key_id: ${OSS_ACCESS_KEY_ID:-}
    access_key_secret: ${OSS_ACCESS_KEY_SECRET:-}
```

## API 使用

```go
import "lychee-go/internal/env"

// 加载 .env 文件
err := env.Load()  // 默认加载 .env
// 或指定路径
err := env.Load(".env", ".env.local")

// 获取环境变量
appName := env.Get("APP_NAME", "default-name")
port := env.GetInt("APP_PORT", 8080)
debug := env.GetBool("APP_DEBUG", false)

// 设置环境变量
env.Set("CUSTOM_VAR", "value")

// 获取所有已加载的环境变量
allVars := env.All()
```

## 初始化流程

1. 在应用启动时调用 `env.Load()` 加载环境变量
2. 然后调用 `config.Init()` 加载配置文件
3. config 模块会自动解析 config.yml 中的 `${ENV_VAR}` 引用

## 优先级

1. 系统环境变量（最高优先级）
2. `.env` 文件中定义的变量
3. config.yml 中的默认值
4. 代码中的默认值（最低优先级）