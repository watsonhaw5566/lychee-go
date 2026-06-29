# websocket · WebSocket 服务

提供 WebSocket 服务支持，包括客户端管理、消息广播和自定义消息处理器。

## 特性

- ✅ 支持 WebSocket 连接管理
- ✅ 支持消息广播（向所有客户端发送）
- ✅ 支持自定义消息处理器注册
- ✅ 支持向单个客户端发送消息
- ✅ 自动处理连接断开和重连
- ✅ 集成到框架 Boot 流程（自动初始化）

## 配置

在 `config.yml` 中配置 WebSocket：

```yaml
websocket:
  enabled: true             # 是否启用 WebSocket（默认启用）
  read_buffer_size: 1024    # 读取缓冲区大小
  write_buffer_size: 1024   # 写入缓冲区大小
  allow_origins: "*"        # 允许的来源（* 表示允许所有）
```

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `enabled` | bool | true | 是否启用 WebSocket |
| `read_buffer_size` | int | 1024 | 读取缓冲区大小（字节） |
| `write_buffer_size` | int | 1024 | 写入缓冲区大小（字节） |
| `allow_origins` | string | "*" | 允许的来源域名，`*` 表示允许所有 |

## API 使用

```go
import "lychee-go/internal/websocket"

// 注册消息处理器（通常在路由初始化时调用）
websocket.RegisterHandler("chat", func(conn *websocket.Conn, message websocket.Message) error {
    var payload struct {
        User    string `json:"user"`
        Content string `json:"content"`
    }
    if err := json.Unmarshal(message.Payload, &payload); err != nil {
        return err
    }
    
    // 广播消息给所有客户端
    return websocket.Broadcast("chat", payload)
})

// 在路由中使用（需要先注册处理器）
r.GET("/ws", websocket.HandleWebSocket)

// 向所有客户端广播消息（可在任意地方调用）
err := websocket.Broadcast("notification", map[string]string{
    "title": "Hello",
    "body":  "Welcome to WebSocket!",
})

// 获取当前在线客户端数量
count := websocket.GetClientCount()
```

## 消息格式

客户端发送和接收的消息都采用 JSON 格式：

```json
{
    "type": "message_type",
    "payload": {...}
}
```

### 错误消息

当处理消息发生错误时，服务端会返回：

```json
{
    "type": "error",
    "payload": {
        "message": "错误描述"
    }
}
```

## 完整示例

```go
package main

import (
    "encoding/json"
    "github.com/gin-gonic/gin"
    "lychee-go/internal/websocket"
)

func main() {
    r := gin.Default()

    // 注册消息处理器
    websocket.RegisterHandler("chat", func(conn *websocket.Conn, msg websocket.Message) error {
        var data struct {
            User    string `json:"user"`
            Content string `json:"content"`
        }
        if err := json.Unmarshal(msg.Payload, &data); err != nil {
            return err
        }
        
        // 广播给所有客户端
        return websocket.Broadcast("chat", data)
    })

    // WebSocket 路由
    r.GET("/ws", websocket.HandleWebSocket)

    r.Run(":8080")
}
```

## 核心 API

| 方法 | 说明 | 参数 | 返回值 |
|------|------|------|--------|
| `HandleWebSocket` | WebSocket 升级处理器 | `w http.ResponseWriter, r *http.Request` | 无 |
| `RegisterHandler` | 注册消息处理器 | `messageType string, handler MessageHandler` | 无 |
| `Broadcast` | 广播消息到所有客户端 | `messageType string, payload interface{}` | `error` |
| `Send` | 向单个客户端发送消息 | `conn *websocket.Conn, messageType string, payload interface{}` | `error` |
| `GetClientCount` | 获取在线客户端数量 | 无 | `int` |
| `Init` | 初始化模块（由 Boot 自动调用） | 无 | 无 |

## 使用场景

1. **实时聊天应用**：通过注册 `chat` 消息处理器实现实时消息传递
2. **实时通知**：使用 `Broadcast` 向所有在线用户推送通知
3. **实时数据更新**：推送股票行情、实时监控数据等
4. **在线状态管理**：通过 `GetClientCount()` 获取在线用户数

## 注意事项

- WebSocket 连接需要在 HTTP 服务器中配置路由
- 建议配合 middleware 进行认证和授权
- 生产环境中应配置 `allow_origins` 为具体域名，禁止使用 `*`
- 模块已集成到框架 Boot 流程，无需手动调用 `Init()`