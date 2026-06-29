# queue · 消息队列

简单的异步任务队列，支持 Redis（生产）和 Memory（开发）两种驱动。

## 用法

### 1. 定义 Job 处理器

```go
// 每个 Job 是一个结构体 + Handle() 方法
type SendEmailJob struct {
    To      string `json:"to"`
    Subject string `json:"subject"`
    Body    string `json:"body"`
}

func (j *SendEmailJob) Handle() error {
    // 实际发送邮件的业务逻辑
    smtp.Send(j.To, j.Subject, j.Body)
    return nil
}

// 在 main 中注册（名字要和 Dispatch 对应）
queue.RegisterJob("send_email", func() queue.Job {
    return &SendEmailJob{}
})
```

### 2. 投递任务

```go
// 投递到默认队列（异步执行，立即返回）
queue.Dispatch("default", "send_email", map[string]interface{}{
    "to":      "user@example.com",
    "subject": "欢迎注册 Lychee-Go",
    "body":    "...",
})

// 自定义最大重试次数
queue.Dispatch("default", "send_email", payload, 5)
```

### 3. 启动 Worker

```go
// 在 main 中启动 worker goroutine
go queue.StartWorker("default")

// 支持多个队列（如：default / high_priority / slow）
go queue.StartWorker("high_priority")
```

## 配置

```yaml
queue:
  driver: memory         # 可选：memory（开发） / redis（生产，分布式）
  prefix: lychee_go_
  default_queue: default
```

## Redis 驱动

当 `driver: redis` 时，队列会使用项目共享的 Redis 连接（由 `cache.GetRedis()` 提供）。

- 支持 **多实例部署**（多台机器可以消费同一个队列）
- 支持 **持久化**（Redis 重启后任务不丢失）
- 队列 key：`{prefix}queue:{name}`

## Memory 驱动

- 最简单，零依赖，适合开发和单机部署
- **注意**：进程重启后未消费的任务会丢失

## 设计要点

- 每个 Job 的 Payload 是 `map[string]interface{}`，被 `json.Marshal` 后投递。
- Job 结构体的字段名必须与 Payload 的 key 对应（通过 `json` tag）。
- `Handle()` 返回 `error` 时，框架会自动重试，直到达到 `maxTries`。
