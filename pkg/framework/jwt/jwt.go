package jwt

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"lychee-go/pkg/framework/config"
	"lychee-go/pkg/framework/logger"
)

// ======== Token 结构 ========

type TokenClaims struct {
	UserID    uint                   `json:"uid"`           // 用户 ID
	LoginType string                 `json:"type"`          // 登录类型：pc / mobile / api
	Device    string                 `json:"device"`        // 设备标识
	IssuedAt  int64                  `json:"iat"`           // 签发时间
	ExpiresAt int64                  `json:"exp"`           // 过期时间
	Extra     map[string]interface{} `json:"ext,omitempty"` // 扩展字段
}

// ======== Token 存储接口 ========

type TokenStore interface {
	Save(token string, claims *TokenClaims) error // 存储 token
	Get(token string) (*TokenClaims, error)       // 获取 token 信息
	Delete(token string) error                    // 删除 token
	DeleteByUserID(userID uint) error             // 删除用户的所有 token（强制下线）
	CountByUserID(userID uint) (int64, error)     // 统计用户 token 数
}

// ======== 内存存储（开发/单机用） ========

type MemoryStore struct {
	tokens map[string]*TokenClaims
	mu     sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{tokens: make(map[string]*TokenClaims)}
}

func (s *MemoryStore) Save(token string, claims *TokenClaims) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[token] = claims
	return nil
}

func (s *MemoryStore) Get(token string) (*TokenClaims, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.tokens[token]
	if !ok {
		return nil, errors.New("token not found")
	}
	return c, nil
}

func (s *MemoryStore) Delete(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, token)
	return nil
}

func (s *MemoryStore) DeleteByUserID(userID uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range s.tokens {
		if v.UserID == userID {
			delete(s.tokens, k)
		}
	}
	return nil
}

func (s *MemoryStore) CountByUserID(userID uint) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var count int64
	for _, v := range s.tokens {
		if v.UserID == userID {
			count++
		}
	}
	return count, nil
}

// ======== JWT 管理器 ========

type Manager struct {
	secret     []byte     // 签名密钥
	expireSec  int64      // 默认过期时间（秒）
	issuer     string     // 签发者
	maxPerUser int        // 单用户最大 token 数
	store      TokenStore // token 存储
}

var manager *Manager
var once sync.Once

// Init 初始化 JWT（全局唯一实例）
func Init() {
	once.Do(func() {
		secretKey := config.GetString("jwt.secret", "lychee-go-secret-key")
		expire := int64(config.GetInt("jwt.ttl", 86400))
		issuer := config.GetString("app.name", "lychee-go")
		maxPerUser := config.GetInt("jwt.max_per_user", 10)

		manager = &Manager{
			secret:     []byte(secretKey),
			expireSec:  expire,
			issuer:     issuer,
			maxPerUser: maxPerUser,
			store:      NewMemoryStore(), // 默认内存存储，可替换为 Redis
		}

		logger.Info("[JWT] Initialized (expire: %ds, max per user: %d)", expire, maxPerUser)
	})
}

// SetStore 替换 token 存储（如切换到 Redis）
func SetStore(store TokenStore) {
	manager.store = store
}

// ======== 签名与验证 ========

func (m *Manager) sign(payload string) string {
	h := sha256.New()
	h.Write([]byte(payload))
	h.Write(m.secret)
	return hex.EncodeToString(h.Sum(nil))
}

func base64Encode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func base64Decode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

// ======== 对外 API ========

// Login 登录并返回 token
// userID: 用户 ID
// loginType: 登录类型（pc/mobile/api 等，方便同账号多端登录）
// extra: 扩展字段（可选）
func Login(userID uint, loginType string, extra ...map[string]interface{}) (string, *TokenClaims, error) {
	if manager == nil {
		Init()
	}

	// 超过最大 token 数时，清除该用户最旧的 token
	count, _ := manager.store.CountByUserID(userID)
	if count >= int64(manager.maxPerUser) {
		_ = manager.store.DeleteByUserID(userID)
	}

	now := time.Now().Unix()
	claims := &TokenClaims{
		UserID:    userID,
		LoginType: loginType,
		Device:    "default",
		IssuedAt:  now,
		ExpiresAt: now + manager.expireSec,
	}
	if len(extra) > 0 && extra[0] != nil {
		claims.Extra = extra[0]
	}

	// 构造 token
	claimsBytes, _ := json.Marshal(claims)
	header := `{"alg":"HS256","typ":"JWT"}`
	payload := base64Encode([]byte(header)) + "." + base64Encode(claimsBytes)
	signature := manager.sign(payload)
	token := payload + "." + signature

	// 存储 token
	if err := manager.store.Save(token, claims); err != nil {
		return "", nil, err
	}

	logger.Info("[JWT] Login: user_id=%d, type=%s", userID, loginType)
	return token, claims, nil
}

// Verify 验证 token
func Verify(token string) (*TokenClaims, error) {
	if manager == nil {
		Init()
	}

	if token == "" {
		return nil, errors.New("token is empty")
	}

	// 检查是否已被手动删除（比如登出/强制下线）
	claims, err := manager.store.Get(token)
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}

	// 校验格式
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	// 校验签名
	payload := parts[0] + "." + parts[1]
	if manager.sign(payload) != parts[2] {
		return nil, errors.New("invalid token signature")
	}

	// 校验过期时间
	if time.Now().Unix() > claims.ExpiresAt {
		_ = manager.store.Delete(token)
		return nil, errors.New("token has expired")
	}

	return claims, nil
}

// Logout 登出（删除 token）
func Logout(token string) error {
	if manager == nil {
		Init()
	}
	logger.Info("[JWT] Logout: token=%s", token[:min(len(token), 20)])
	return manager.store.Delete(token)
}

// KickOut 强制下线用户（删除该用户所有 token）
func KickOut(userID uint) error {
	if manager == nil {
		Init()
	}
	logger.Info("[JWT] KickOut: user_id=%d", userID)
	return manager.store.DeleteByUserID(userID)
}

// Refresh 刷新 token（延长有效期）
func Refresh(token string) (string, *TokenClaims, error) {
	claims, err := Verify(token)
	if err != nil {
		return "", nil, err
	}

	// 删除旧 token
	_ = manager.store.Delete(token)

	// 重新登录，使用相同的用户 ID
	return Login(claims.UserID, claims.LoginType, claims.Extra)
}

// GetTokenTTL 获取 token 剩余有效期（秒）
func GetTokenTTL(token string) int64 {
	claims, err := Verify(token)
	if err != nil {
		return 0
	}
	return claims.ExpiresAt - time.Now().Unix()
}

// IsLoggedIn 检查用户是否在线（至少有一个有效 token）
func IsLoggedIn(userID uint) bool {
	if manager == nil {
		Init()
	}
	count, _ := manager.store.CountByUserID(userID)
	return count > 0
}

// ======== 辅助函数 ========

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ======== 扩展：常用 Login 方法 ========

// LoginByID 便捷登录（只需要用户 ID）
func LoginByID(userID uint) (string, error) {
	token, _, err := Login(userID, "api", nil)
	return token, err
}

// LoginWithExtra 登录并携带扩展字段
func LoginWithExtra(userID uint, extra map[string]interface{}) (string, error) {
	token, _, err := Login(userID, "api", extra)
	return token, err
}

// GetUserIDFromToken 从 token 中提取用户 ID
func GetUserIDFromToken(token string) (uint, error) {
	claims, err := Verify(token)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

// GetExtraFromToken 从 token 中获取扩展字段
func GetExtraFromToken(token, key string) (interface{}, error) {
	claims, err := Verify(token)
	if err != nil {
		return nil, err
	}
	if claims.Extra == nil {
		return nil, fmt.Errorf("extra field not found: %s", key)
	}
	v, ok := claims.Extra[key]
	if !ok {
		return nil, fmt.Errorf("extra field not found: %s", key)
	}
	return v, nil
}
