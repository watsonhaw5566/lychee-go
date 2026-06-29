# cookie · Cookie 管理

完整的 Cookie 操作，支持签名 Cookie（防篡改）、JSON 存取、闪现数据。

## 基本用法

```go
import "lychee-go/internal/cookie"

// ---------- 设置 ----------
cookie.Set(c, "username", "alice")                          // 默认 1 天
cookie.Set(c, "theme", "dark", 7*24*time.Hour)             // 自定义过期时间

// ---------- 读取 ----------
username, ok := cookie.Get(c, "username")                  // (value, exists)
username := cookie.GetOr(c, "username", "guest")            // 带默认值

// ---------- 检查 / 删除 ----------
if cookie.Has(c, "username") { ... }
cookie.Delete(c, "username")
cookie.ClearAll(c)                                          // 删除所有 Cookie
```

## 签名 Cookie（**防篡改，推荐**）

使用 `HMAC-SHA256` 对 Cookie 值签名，防止用户在浏览器中手动修改。

```go
// 设置
cookie.SetSigned(c, "user_id", "12345")

// 读取（自动验证签名）
userID, ok := cookie.GetSigned(c, "user_id")
if !ok {
    // Cookie 不存在或签名被篡改 → 已自动删除
    return "请重新登录"
}

// 带默认值
userID := cookie.GetSignedOr(c, "user_id", "0")
```

**适合存：** 用户 ID、角色、语言偏好、主题、记住我 Token 等。

**签名格式：** `base64(value) . hmac_sha256_signature`

## JSON Cookie（对象级存取）

```go
type Preferences struct {
    Theme    string `json:"theme"`
    Language string `json:"language"`
}

// 存
prefs := Preferences{Theme: "dark", Language: "zh-CN"}
cookie.SetJSON(c, "prefs", prefs)

// 读
var prefs Preferences
if cookie.GetJSON(c, "prefs", &prefs) {
    fmt.Println(prefs.Theme)
}

// 签名版（防篡改 + 对象级）
cookie.SetSignedJSON(c, "prefs", prefs)
cookie.GetSignedJSON(c, "prefs", &prefs)
```

## Flash Cookie（**读一次即删除**）

适合表单提交后跳转回显错误消息：

```go
// POST /login — 登录失败
cookie.Flash(c, "error", "用户名或密码错误")
c.Redirect(302, "/login")

// GET /login — 读取后自动删除
if msg, ok := cookie.GetFlash(c, "error"); ok {
    // 显示错误消息
}

// 也支持对象
cookie.FlashJSON(c, "form_data", previousInput)
cookie.GetFlashJSON(c, "form_data", &input)
```

## 便捷方法

```go
// "记住我" — 30 天长期 Cookie
cookie.Remember(c, "remember_token", "abcd1234")
cookie.RememberSigned(c, "user_id", "12345")

// 临时 Cookie — 浏览器关闭即失效（会话级）
cookie.Temporary(c, "temp_data", "some value")

// 语言偏好（自动签名 + 1 年）
cookie.SetLanguage(c, "zh-CN")
lang := cookie.GetLanguage(c)         // "zh-CN"

// 主题偏好
cookie.SetTheme(c, "dark")
theme := cookie.GetTheme(c)           // "dark"
```

## 纯 net/http 场景（无需 Gin）

```go
func handler(w http.ResponseWriter, r *http.Request) {
    cookie.RawSet(w, "name", "value")
    cookie.RawSetSigned(w, "user_id", "12345")

    value, ok := cookie.RawGet(r, "name")
    userID, ok := cookie.RawGetSigned(r, "user_id")

    cookie.RawDelete(w, "name")
}
```

## 配置

```yaml
cookie:
  secret: "your-strong-secret-key"   # 签名密钥（务必保密且够长）
  domain: ""                          # Cookie 作用域（空 = 当前域名）
  path: "/"                           # Cookie 路径
  max_age: 86400                      # 默认过期时间（秒，1 天）
  secure: false                       # HTTPS 环境请设 true
  httponly: true                      # 禁止 JS 访问（防 XSS）
  samesite: "lax"                     # lax / strict / none（防 CSRF）
  prefix: ""                          # Cookie 名前缀（可选）
```

## 安全建议

**生产环境：**
- ✅ `secure: true`（仅 HTTPS 传输）
- ✅ `httponly: true`（防 XSS 读取敏感 Cookie）
- ✅ `samesite: "lax"` 或 `"strict"`（防 CSRF）
- ✅ `secret` 使用至少 32 位随机字符串
- ✅ 用户 ID、角色等需要可信的字段使用 `SetSigned` / `GetSigned`
