# logger · 日志系统

基于 [zap](https://github.com/uber-go/zap) 封装的高性能结构化日志。

## 用法

```go
import "lychee-go/internal/logger"

// 由 boot.Boot() 自动初始化，之后直接使用
logger.Info("User login: %s, role: %s", username, role)
logger.Warn("Rate limit hit: %s", clientIP)
logger.Error("Database error: %v", err)
logger.Debug("Request body: %s", body)
```

## 日志级别

| 级别 | 用途 |
|------|------|
| `Debug` | 调试信息，生产环境默认关闭 |
| `Info` | 常规业务信息 |
| `Warn` | 警告（非致命错误） |
| `Error` | 错误（需要关注） |

## 配置

```yaml
log:
  level: debug                 # 日志级别
  filename: runtime/logs/app.log
  max_size: 100                # 单文件最大 MB
  max_backups: 30              # 历史文件保留数
  max_age: 7                   # 历史文件保留天数
  compress: true               # 是否 gzip 压缩历史文件
```

## 日志格式

- **控制台**：彩色文本，方便开发调试
- **文件**：JSON 格式，便于日志采集系统（ELK / Loki / Splunk）解析

每一条日志都包含：时间戳、级别、调用文件/行号、消息内容。
