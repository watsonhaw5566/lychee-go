# SaToken 模块

基于 Think-SaToken PHP 实现参考，在 Go 中实现的 SaToken 权限认证框架。

## 功能特性

- **登录认证**：生成 Token 并存储到缓存
- **登出功能**：删除 Token 并清理登录映射
- **并发登录控制**：支持同一账号多地登录，可配置最大登录数量
- **滑动续期**：访问自动延长 Token 有效期
- **Token 验证**：格式校验、有效期检查
- **强制下线**：支持按用户 ID 或 Token 踢出
- **扩展字段**：支持在 Token 中存储自定义数据

## 配置项

在 `config/config.yml` 中配置：

```yaml
satoken:
  token_name: ''           # 自定义 Token name 名称
  timeout: 86400          # Token 有效期，单位秒（默认1天）
  is_concurrent: true     # 是否允许同一账号多地登录
  max_login_count: 10     # 同一账号最大登录数量
  auto_renew: true        # 是否启用滑动续期
  renew_threshold: 0.3    # 滑动续期阈值（剩余时间低于此比例触发续期）
```

## 使用示例

```go
package main

import (
    "lychee-go/internal/satoken"
)

func main() {
    // 初始化（通常在启动时调用）
    satoken.Init()

    // 登录
    token, err := satoken.Login(1001, map[string]interface{}{"name": "张三"})
    if err != nil {
        // 处理错误
    }

    // 检查登录状态
    isLogin := satoken.IsLogin(token)

    // 获取当前登录用户ID
    loginID, err := satoken.GetCurrentLoginId(token)

    // 获取 Token 信息
    info, err := satoken.GetTokenInfo(token)

    // 登出
    satoken.Logout(token)

    // 强制踢出用户
    satoken.Kickout(1001)
}
```

## 缓存依赖

本模块依赖 `internal/cache` 模块，需要确保 Redis 缓存已正确配置和初始化。

## API 列表

| 方法 | 说明 | 参数 | 返回值 |
|------|------|------|--------|
| `Login` | 用户登录 | `loginID int64, extra ...map[string]interface{}` | `(token string, err error)` |
| `Logout` | 用户登出 | `token ...string` | `bool` |
| `IsLogin` | 检查是否登录 | `token ...string` | `bool` |
| `CheckLogin` | 检查登录（抛出异常） | `token ...string` | `error` |
| `GetCurrentLoginId` | 获取登录用户 ID | `token ...string` | `(int64, error)` |
| `GetTokenInfo` | 获取 Token 信息 | `token ...string` | `(*TokenInfo, error)` |
| `GetExtra` | 获取扩展字段 | `token ...string` | `(map[string]interface{}, error)` |
| `SetExtra` | 设置扩展字段 | `extra map[string]interface{}, token ...string` | `bool` |
| `GetTokenExpireTime` | 获取过期时间戳 | `token ...string` | `int64` |
| `GetTokenRemainingTime` | 获取剩余秒数 | `token ...string` | `int64` |
| `Kickout` | 强制踢出用户 | `loginID int64` | `bool` |
| `KickoutByToken` | 强制踢出 Token | `token string` | `bool` |

## 错误定义

| 错误变量 | 错误信息 |
|----------|----------|
| `ErrNotLogin` | 未提供token |
| `ErrTokenInvalid` | 无效的token |
| `ErrTokenFormat` | 无效的token格式 |
| `ErrTokenInfo` | token信息不完整 |
| `ErrTokenExpired` | token已过期 |

## Token 信息结构

```go
type TokenInfo struct {
    LoginID    int64                  // 用户登录ID
    CreateTime int64                  // 创建时间戳
    ExpireTime int64                  // 过期时间戳
    Extra      map[string]interface{} // 扩展字段
}
```

## 核心流程

### 登录流程

1. 生成 UUID v4 格式的 Token
2. 获取分布式锁防止并发冲突
3. 获取用户已有的 Token 列表
4. 添加新 Token 并清理无效 Token
5. 如果超过最大登录数，移除最早的 Token
6. 存储 Token 列表和 Token 信息
7. 释放分布式锁

### 登出流程

1. 解析 Token 获取用户信息
2. 获取分布式锁
3. 从用户 Token 列表中移除该 Token
4. 删除 Token 信息
5. 释放分布式锁

### 滑动续期

1. 检查是否开启自动续期
2. 判断剩余时间是否低于阈值
3. 更新 Token 过期时间
4. 同步更新用户 Token 列表的过期时间