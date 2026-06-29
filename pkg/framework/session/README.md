# session · 会话管理

基于 Cookie 的 Session，带 Key 前缀和自动过期清理。

## 用法

```go
import "lychee-go/internal/session"

// ---------- 创建新 Session ----------
sess, err := session.Create()
sessionID := sess.ID        // 把这个 ID 放到 Cookie 里返回给前端

// ---------- 读写 ----------
sess.Set("user_id", 123)
sess.Set("username", "alice")

userID := sess.GetUint("user_id")
username := sess.GetString("username")

if sess.Has("login_time") { ... }

sess.Delete("temp_data")        // 删除单个键
sess.Clear()                    // 清空所有数据（保留 Session）

// ---------- 读取已有 Session ----------
sess, err := session.Get(sessionID)
if err != nil {
    // Session 不存在或已过期
}

// ---------- 销毁 Session ----------
session.Delete(sessionID)
```

## 类型安全的 Get

| 方法 | 返回类型 | 示例 |
|------|----------|------|
| `Get(key)` | `interface{}` | `sess.Get("user_id")` |
| `GetString(key)` | `string` | `sess.GetString("username")` |
| `GetInt(key)` | `int` | `sess.GetInt("age")` |
| `GetUint(key)` | `uint` | `sess.GetUint("user_id")` |
| `GetBool(key)` | `bool` | `sess.GetBool("logged_in")` |
| `All()` | `map[string]interface{}` | `sess.All()` |

## 闪现数据（Flash）

适合"登录失败提示"这类 **读一次即消失** 的场景：

```go
// POST /login — 失败后跳转
sess.Flash("error", "用户名或密码错误")
c.Redirect(302, "/login")

// GET /login — 读取后自动删除
msg := sess.GetFlash("error")
```

## 配置

```yaml
session:
  ttl: 7200          # Session 过期时间（秒，2 小时）
  name: lychee_session   # Cookie 名
```

## 自动过期

后台 goroutine 每 10 分钟清理一次过期 Session，无需手动维护。

## 与 Cookie 模块配合

```go
// 登录成功时
sess, _ := session.Create()
sess.Set("user_id", user.ID)
cookie.SetSigned(c, "session_id", sess.ID, 2*time.Hour)

// 后续请求验证
sessionID, ok := cookie.GetSigned(c, "session_id")
if !ok { return }
sess, err := session.Get(sessionID)
if err != nil { return }
userID := sess.GetUint("user_id")
```
