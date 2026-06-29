package cookie

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"lychee-go/pkg/framework/config"
	"lychee-go/pkg/framework/logger"

	"github.com/gin-gonic/gin"
)

// ======== Cookie 配置 ========

type Config struct {
	Secret   string        // 签名密钥（用于 Signed Cookie）
	Domain   string        // Cookie 作用域
	Path     string        // Cookie 路径
	MaxAge   time.Duration // 默认过期时间
	Secure   bool          // 是否仅 HTTPS 传输
	HTTPOnly bool          // 是否禁止 JS 访问
	SameSite string        // SameSite 策略：lax / strict / none
	Prefix   string        // Cookie 名称前缀
}

var (
	defaultConfig *Config
	once          sync.Once
)

// ======== 初始化 ========

func Init() {
	once.Do(func() {
		maxAgeSec := config.GetInt("cookie.max_age", 86400)

		defaultConfig = &Config{
			Secret:   config.GetString("cookie.secret", "lychee-go-cookie-secret-change-it"),
			Domain:   config.GetString("cookie.domain", ""),
			Path:     config.GetString("cookie.path", "/"),
			MaxAge:   time.Duration(maxAgeSec) * time.Second,
			Secure:   config.GetBool("cookie.secure", false),
			HTTPOnly: config.GetBool("cookie.httponly", true),
			SameSite: config.GetString("cookie.samesite", "lax"),
			Prefix:   config.GetString("cookie.prefix", ""),
		}

		logger.Info("[Cookie] Initialized (path: %s, max_age: %ds, httponly: %v, samesite: %s)",
			defaultConfig.Path, maxAgeSec, defaultConfig.HTTPOnly, defaultConfig.SameSite)
	})
}

// GetConfig 获取当前配置（需要自定义时使用）
func GetConfig() *Config {
	if defaultConfig == nil {
		Init()
	}
	return defaultConfig
}

// ======== 签名算法 ========

// sign 生成签名（HMAC-SHA256）
func sign(data, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

// signCookieValue 对 Cookie 值签名，格式：base64(value) + "." + signature
func signCookieValue(value, secret string) string {
	encoded := base64.RawURLEncoding.EncodeToString([]byte(value))
	signature := sign(encoded, secret)
	return encoded + "." + signature
}

// verifyCookieValue 验证签名并返回原始值
func verifyCookieValue(signed, secret string) (string, bool) {
	parts := strings.Split(signed, ".")
	if len(parts) != 2 {
		return "", false
	}
	encoded, signature := parts[0], parts[1]

	// 重新计算签名并比对（防篡改）
	if sign(encoded, secret) != signature {
		return "", false
	}

	// 解码原始值
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", false
	}
	return string(data), true
}

// ======== SameSite 解析 ========

func parseSameSite(s string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	case "lax", "":
		return http.SameSiteLaxMode
	default:
		return http.SameSiteDefaultMode
	}
}

// ======== 内部工具：设置 Cookie ========

func buildCookie(name, value string, maxAge time.Duration, cfg *Config) *http.Cookie {
	if cfg.Prefix != "" && !strings.HasPrefix(name, cfg.Prefix) {
		name = cfg.Prefix + name
	}

	maxAgeSec := 0
	if maxAge > 0 {
		maxAgeSec = int(maxAge.Seconds())
	}

	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     cfg.Path,
		Domain:   cfg.Domain,
		MaxAge:   maxAgeSec,
		Secure:   cfg.Secure,
		HttpOnly: cfg.HTTPOnly,
		SameSite: parseSameSite(cfg.SameSite),
	}
}

// ======== Gin 上下文操作：普通 Cookie ========

// Set 设置普通 Cookie
func Set(c *gin.Context, name, value string, options ...time.Duration) {
	cfg := GetConfig()
	maxAge := cfg.MaxAge
	if len(options) > 0 {
		maxAge = options[0]
	}
	cookie := buildCookie(name, value, maxAge, cfg)
	http.SetCookie(c.Writer, cookie)
}

// Get 获取普通 Cookie
func Get(c *gin.Context, name string) (string, bool) {
	cfg := GetConfig()
	fullName := name
	if cfg.Prefix != "" {
		fullName = cfg.Prefix + name
	}
	cookie, err := c.Request.Cookie(fullName)
	if err != nil {
		return "", false
	}
	return cookie.Value, true
}

// GetOr 获取 Cookie，不存在返回默认值
func GetOr(c *gin.Context, name, defaultValue string) string {
	if v, ok := Get(c, name); ok {
		return v
	}
	return defaultValue
}

// Delete 删除 Cookie（设置过期时间为过去）
func Delete(c *gin.Context, name string) {
	cfg := GetConfig()
	fullName := name
	if cfg.Prefix != "" {
		fullName = cfg.Prefix + name
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     fullName,
		Value:    "",
		Path:     cfg.Path,
		Domain:   cfg.Domain,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		Secure:   cfg.Secure,
		HttpOnly: cfg.HTTPOnly,
	})
}

// Has 检查 Cookie 是否存在
func Has(c *gin.Context, name string) bool {
	_, ok := Get(c, name)
	return ok
}

// ======== Gin 上下文操作：签名 Cookie（防篡改） ========

// SetSigned 设置签名 Cookie（值会被 HMAC-SHA256 签名，防止用户篡改）
// 适合存：用户 ID、角色、语言偏好等非敏感但需可信的数据
func SetSigned(c *gin.Context, name, value string, options ...time.Duration) {
	cfg := GetConfig()
	maxAge := cfg.MaxAge
	if len(options) > 0 {
		maxAge = options[0]
	}
	signed := signCookieValue(value, cfg.Secret)
	cookie := buildCookie(name, signed, maxAge, cfg)
	http.SetCookie(c.Writer, cookie)
}

// GetSigned 获取签名 Cookie 并验证签名
func GetSigned(c *gin.Context, name string) (string, bool) {
	cfg := GetConfig()
	fullName := name
	if cfg.Prefix != "" {
		fullName = cfg.Prefix + name
	}
	cookie, err := c.Request.Cookie(fullName)
	if err != nil {
		return "", false
	}
	value, ok := verifyCookieValue(cookie.Value, cfg.Secret)
	if !ok {
		logger.Warn("[Cookie] Invalid signature on: %s", fullName)
		// 签名被篡改，删除该 Cookie
		Delete(c, name)
		return "", false
	}
	return value, true
}

// GetSignedOr 获取签名 Cookie，不存在或验证失败返回默认值
func GetSignedOr(c *gin.Context, name, defaultValue string) string {
	if v, ok := GetSigned(c, name); ok {
		return v
	}
	return defaultValue
}

// ======== JSON Cookie：对象级存取 ========

// SetJSON 以 JSON 格式存对象到 Cookie
func SetJSON(c *gin.Context, name string, value interface{}, options ...time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	Set(c, name, string(data), options...)
	return nil
}

// GetJSON 从 Cookie 中读取 JSON 并解析到目标对象
func GetJSON(c *gin.Context, name string, target interface{}) bool {
	value, ok := Get(c, name)
	if !ok {
		return false
	}
	return json.Unmarshal([]byte(value), target) == nil
}

// SetSignedJSON 签名 JSON Cookie（防篡改 + 对象级）
func SetSignedJSON(c *gin.Context, name string, value interface{}, options ...time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	SetSigned(c, name, string(data), options...)
	return nil
}

// GetSignedJSON 读取签名 JSON Cookie
func GetSignedJSON(c *gin.Context, name string, target interface{}) bool {
	value, ok := GetSigned(c, name)
	if !ok {
		return false
	}
	return json.Unmarshal([]byte(value), target) == nil
}

// ======== Flash Cookie：只读一次（适合表单提交后的提示） ========

// Flash 设置闪现数据（下次读取后自动删除）
func Flash(c *gin.Context, name, value string, options ...time.Duration) {
	Set(c, "__flash__"+name, value, options...)
}

// GetFlash 获取闪现数据（读取后删除）
func GetFlash(c *gin.Context, name string) (string, bool) {
	value, ok := Get(c, "__flash__"+name)
	if ok {
		Delete(c, "__flash__"+name)
	}
	return value, ok
}

// FlashJSON 设置闪现 JSON
func FlashJSON(c *gin.Context, name string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	Flash(c, name, string(data))
	return nil
}

// GetFlashJSON 获取闪现 JSON
func GetFlashJSON(c *gin.Context, name string, target interface{}) bool {
	value, ok := GetFlash(c, name)
	if !ok {
		return false
	}
	return json.Unmarshal([]byte(value), target) == nil
}

// ======== 纯 http.ResponseWriter / *http.Request 操作（无需 Gin 也能用） ========

// RawSet 直接用 http.ResponseWriter 设置 Cookie
func RawSet(w http.ResponseWriter, name, value string, options ...time.Duration) {
	cfg := GetConfig()
	maxAge := cfg.MaxAge
	if len(options) > 0 {
		maxAge = options[0]
	}
	http.SetCookie(w, buildCookie(name, value, maxAge, cfg))
}

// RawGet 从 *http.Request 读取 Cookie
func RawGet(r *http.Request, name string) (string, bool) {
	cfg := GetConfig()
	fullName := name
	if cfg.Prefix != "" {
		fullName = cfg.Prefix + name
	}
	cookie, err := r.Cookie(fullName)
	if err != nil {
		return "", false
	}
	return cookie.Value, true
}

// RawDelete 删除 Cookie
func RawDelete(w http.ResponseWriter, name string) {
	cfg := GetConfig()
	fullName := name
	if cfg.Prefix != "" {
		fullName = cfg.Prefix + name
	}
	http.SetCookie(w, &http.Cookie{
		Name:     fullName,
		Value:    "",
		Path:     cfg.Path,
		Domain:   cfg.Domain,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		Secure:   cfg.Secure,
		HttpOnly: cfg.HTTPOnly,
	})
}

// RawSetSigned 纯 http 版签名 Cookie
func RawSetSigned(w http.ResponseWriter, name, value string, options ...time.Duration) {
	cfg := GetConfig()
	maxAge := cfg.MaxAge
	if len(options) > 0 {
		maxAge = options[0]
	}
	signed := signCookieValue(value, cfg.Secret)
	http.SetCookie(w, buildCookie(name, signed, maxAge, cfg))
}

// RawGetSigned 纯 http 版读取签名 Cookie
func RawGetSigned(r *http.Request, name string) (string, bool) {
	cfg := GetConfig()
	fullName := name
	if cfg.Prefix != "" {
		fullName = cfg.Prefix + name
	}
	cookie, err := r.Cookie(fullName)
	if err != nil {
		return "", false
	}
	value, ok := verifyCookieValue(cookie.Value, cfg.Secret)
	if !ok {
		logger.Warn("[Cookie] Invalid signature on: %s", fullName)
	}
	return value, ok
}

// ======== 便捷：常见场景 ========

// Remember 设置长期 Cookie（30 天，适合"记住我"）
func Remember(c *gin.Context, name, value string) {
	Set(c, name, value, 30*24*time.Hour)
}

// RememberSigned 长期签名 Cookie
func RememberSigned(c *gin.Context, name, value string) {
	SetSigned(c, name, value, 30*24*time.Hour)
}

// Temporary 临时 Cookie（会话级，浏览器关闭即失效）
func Temporary(c *gin.Context, name, value string) {
	cfg := GetConfig()
	cookie := buildCookie(name, value, 0, cfg)
	cookie.MaxAge = 0
	cookie.Expires = time.Time{}
	http.SetCookie(c.Writer, cookie)
}

// SetLanguage 设置语言偏好
func SetLanguage(c *gin.Context, lang string) {
	SetSigned(c, "lang", lang, 365*24*time.Hour)
}

// GetLanguage 获取语言偏好
func GetLanguage(c *gin.Context) string {
	return GetSignedOr(c, "lang", "zh-CN")
}

// SetTheme 设置主题
func SetTheme(c *gin.Context, theme string) {
	SetSigned(c, "theme", theme, 365*24*time.Hour)
}

// GetTheme 获取主题
func GetTheme(c *gin.Context) string {
	return GetSignedOr(c, "theme", "light")
}

// ClearAll 清除所有 Cookie（遍历当前请求中的 Cookie 并删除）
func ClearAll(c *gin.Context) {
	for _, cookie := range c.Request.Cookies() {
		Delete(c, cookie.Name)
	}
}

// ======== Gin 中间件：便捷注入 ========

// Middleware 可选的中间件（目前无特殊逻辑，预留未来扩展）
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
