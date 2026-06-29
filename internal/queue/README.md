# queue · 消息队列

简单的异步任务队列，支持以下驱动：
- **Memory**：开发测试用，零依赖
- **Redis**：生产环境，支持分布式部署
- **TDCMQ**：腾讯云消息队列，支持高可用分布式场景

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
  driver: memory         # 可选：memory（开发） / redis（生产，分布式） / tdcmq（腾讯云）
  prefix: lychee_go_
  default_queue: default
  
  # TDCMQ 配置（driver: tdcmq 时生效）
  tdcmq:
    url: ""              # 可选，手动配置完整 URL，优先级最高
    network: public      # 网络类型：public（公网）/ private（内网）
    secret_id: ${TDCMQ_SECRET_ID}
    secret_key: ${TDCMQ_SECRET_KEY}
    region: ap-guangzhou
```

## Redis 驱动

当 `driver: redis` 时，队列会使用项目共享的 Redis 连接（由 `cache.GetRedis()` 提供）。

- 支持 **多实例部署**（多台机器可以消费同一个队列）
- 支持 **持久化**（Redis 重启后任务不丢失）
- 队列 key：`{prefix}queue:{name}`

## Memory 驱动

- 最简单，零依赖，适合开发和单机部署
- **注意**：进程重启后未消费的任务会丢失

## TDCMQ 驱动

当 `driver: tdcmq` 时，队列会使用腾讯云 TDCMQ（消息队列 CMQ）服务。

### 特性

- **高可用**：腾讯云托管服务，多可用区部署
- **分布式**：支持多实例并发消费
- **持久化**：消息持久化存储，不丢失
- **弹性伸缩**：根据业务流量自动扩容

### 配置说明

| 配置项 | 必填 | 默认值 | 说明 |
|--------|------|--------|------|
| `url` | 否 | 自动生成 | CMQ 服务端点 URL，优先级最高 |
| `network` | 否 | public | 网络类型：`public`（公网）/ `private`（内网） |
| `secret_id` | 是 | - | 腾讯云 API SecretId |
| `secret_key` | 是 | - | 腾讯云 API SecretKey |
| `region` | 否 | ap-guangzhou | 地域，如：ap-guangzhou, ap-beijing |

### URL 自动生成规则

根据 `network` 和 `region` 自动生成对应的 API 地址：

| network | region | 生成的 URL |
|---------|--------|------------|
| public | ap-guangzhou | `https://cmq-gz.public.tencenttdmq.com` |
| private | ap-guangzhou | `http://gz.mqadapter.cmq.tencentytun.com` |
| public | ap-beijing | `https://cmq-bj.public.tencenttdmq.com` |
| private | ap-beijing | `http://bj.mqadapter.cmq.tencentytun.com` |
| public | ap-shanghai | `https://cmq-sh.public.tencenttdmq.com` |
| private | ap-shanghai | `http://sh.mqadapter.cmq.tencentytun.com` |

**支持的地域**：`ap-guangzhou`(gz), `ap-beijing`(bj), `ap-shanghai`(sh), `ap-shenzhen`(sz), `ap-hongkong`(hk), `ap-tokyo`(jp), `ap-seoul`(kr), `eu-frankfurt`(de), `us-east`(us)

**手动配置**：如果自动生成的地址不符合需求，可以直接配置 `url` 字段，优先级最高。

### 使用方式

1. **在腾讯云控制台创建队列**：
   - 登录 [腾讯云控制台](https://console.cloud.tencent.com/tdmq)
   - 创建 CMQ 队列，记录队列名称

2. **配置环境变量**：
   ```bash
   export TDCMQ_SECRET_ID=your-secret-id
   export TDCMQ_SECRET_KEY=your-secret-key
   ```

3. **配置驱动**：
   ```yaml
   queue:
     driver: tdcmq
     tdcmq:
       region: ap-guangzhou
   ```

### 注意事项

- 需要提前在腾讯云控制台创建好队列
- 队列名称需要与代码中使用的队列名一致
- 需要确保腾讯云账号具有 CMQ 相关权限

## 设计要点

- 每个 Job 的 Payload 是 `map[string]interface{}`，被 `json.Marshal` 后投递。
- Job 结构体的字段名必须与 Payload 的 key 对应（通过 `json` tag）。
- `Handle()` 返回 `error` 时，框架会自动重试，直到达到 `maxTries`。