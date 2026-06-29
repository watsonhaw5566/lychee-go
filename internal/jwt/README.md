# jwt · Token 鉴权

轻量级的 Token 管理，类似 Sa-Token 的 API，支持：
- 签发 / 验证 Token
- 登出 / 强制下线
- Token 过期 / 多端登录数量限制
- 扩展字段携带

## 用法

```go
import "lychee-go/internal/jwt"

// ---------- 登录 ----------
token, claims, err := jwt.Login(userID, "api",
    map[string]interface{}{"username": "alice", "role": "admin"})
// token    : 字符串，发给前端
// claims   : 包含 userID、登录类型、签发时间等
// userID 为 uint，第二个参数为登录类型（区分 web / api / mobile 多端）

// ---------- 验证 Token ----------
claims, err := jwt.Verify(token)
if err != nil {
    // Token 无效 / 过期 / 被登出
    return
}
userID := claims.UserID

// ---------- 登出（删除当前 Token） ----------
jwt.Logout(token)

// ---------- 强制下线用户（删除该用户所有 Token） ----------
jwt.KickOut(userID)

// ---------- 刷新 Token ----------
newToken, newClaims, err := jwt.Refresh(token)

// ---------- 查询 ----------
jwt.IsLoggedIn(userID)         // 该用户是否在线（至少有一个有效 Token）
jwt.GetTokenTTL(token)         // Token 剩余有效时间（秒）
```

## 便捷方法

```go
// 只需要用户 ID 的登录
token, err := jwt.LoginByID(userID)

// 携带扩展字段
token, err := jwt.LoginWithExtra(userID,
    map[string]interface{}{"tenant": "acme"})

// 从 Token 中提取用户 ID
userID, err := jwt.GetUserIDFromToken(token)

// 从 Token 中读取扩展字段
tenant, err := jwt.GetExtraFromToken(token, "tenant")
```

## 配置

```yaml
auth:
  jwt_secret: "your-strong-secret-key"  # HMAC 签名密钥，务必保密
  jwt_ttl: 86400                         # Token 过期时间（秒，1 天）
  max_per_user: 10                       # 每个用户最多同时在线 Token 数
```

## Token 格式

```
base64(header) . base64(claims) . hmac_sha256_signature
```

- **Header**：`{"alg":"HS256","typ":"JWT"}`
- **Claims**：用户 ID、登录类型、签发时间、过期时间、扩展字段
- **Signature**：对前两部分的 HMAC-SHA256 签名（防止篡改）

## 存储

默认使用 **内存存储**（`memoryStore`），适合单机部署。

切换到 Redis（分布式部署）：

```go
// 实现 Store 接口
type RedisStore struct{ ... }
func (s *RedisStore) Save(token string, claims *TokenClaims) error { ... }
func (s *RedisStore) Get(token string) (*TokenClaims, error) { ... }
func (s *RedisStore) Delete(token string) error { ... }
func (s *RedisStore) DeleteByUserID(userID uint) error { ... }
func (s *RedisStore) CountByUserID(userID uint) (int64, error) { ... }

// 注册
jwt.SetStore(&RedisStore{...})
```
