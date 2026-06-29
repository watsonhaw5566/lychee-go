package satoken

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

	"lychee-go/internal/cache"
	"lychee-go/internal/config"
	"lychee-go/internal/logger"

	"github.com/google/uuid"
)

var (
	ErrNotLogin     = errors.New("未提供token")
	ErrTokenInvalid = errors.New("无效的token")
	ErrTokenFormat  = errors.New("无效的token格式")
	ErrTokenInfo    = errors.New("token信息不完整")
	ErrTokenExpired = errors.New("token已过期")
)

type TokenInfo struct {
	LoginID    int64                  `json:"loginId"`     // 用户登录ID
	CreateTime int64                  `json:"create_time"` // 创建时间
	ExpireTime int64                  `json:"expire_time"` // 过期时间
	Extra      map[string]interface{} `json:"extra"`       // 额外自定义内容
}

type Config struct {
	TokenName      string  `json:"token_name"`      // 自定义 Token name 名称
	Timeout        int64   `json:"timeout"`         // Token 有效期，单位秒
	IsConcurrent   bool    `json:"is_concurrent"`   // 是否允许同一账号多地登录
	MaxLoginCount  int     `json:"max_login_count"` // 同一账号最大登录数量
	AutoRenew      bool    `json:"auto_renew"`      // 是否启用滑动续期
	RenewThreshold float64 `json:"renew_threshold"` // 滑动续期阈值 (0~1)
}

type SaToken struct {
	config        Config
	isRedisDriver bool
	mu            sync.RWMutex
	memoryLocks   sync.Map
}

var instance *SaToken
var once sync.Once

func Init() {
	once.Do(func() {
		cfg := Config{
			TokenName:      config.GetString("satoken.token_name", ""),
			Timeout:        int64(config.GetInt("satoken.timeout", 86400)),
			IsConcurrent:   config.GetBool("satoken.is_concurrent", true),
			MaxLoginCount:  config.GetInt("satoken.max_login_count", 10),
			AutoRenew:      config.GetBool("satoken.auto_renew", true),
			RenewThreshold: config.GetFloat64("satoken.renew_threshold", 0.3),
		}

		if cfg.RenewThreshold <= 0 || cfg.RenewThreshold > 1 {
			cfg.RenewThreshold = 0.3
		}

		driver := config.GetString("cache.driver", "memory")
		isRedis := driver == "redis"

		instance = &SaToken{
			config:        cfg,
			isRedisDriver: isRedis,
		}
		logger.Info("[SaToken] Initialized (driver: %s, timeout: %ds, max_login: %d, auto_renew: %v)",
			driver, cfg.Timeout, cfg.MaxLoginCount, cfg.AutoRenew)
	})
}

func getInstance() *SaToken {
	if instance == nil {
		Init()
	}
	return instance
}

func tokenKey(token string) string {
	return "satoken:token:" + token
}

func loginIDKey(loginID int64) string {
	return "satoken:loginId:" + fmt.Sprintf("%d", loginID)
}

func lockKey(key string) string {
	return "satoken:lock:" + key
}

func (s *SaToken) validateTokenFormat(token string) bool {
	if len(token) != 36 {
		return false
	}
	pattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	return pattern.MatchString(token)
}

func (s *SaToken) createToken() string {
	return uuid.New().String()
}

func (s *SaToken) resolveMaxLoginCount() int {
	if !s.config.IsConcurrent {
		return 1
	}
	if s.config.MaxLoginCount <= 0 {
		return 1
	}
	return s.config.MaxLoginCount
}

func (s *SaToken) getMinRemainingTime(tokenList []string) int64 {
	min := s.config.Timeout
	now := time.Now().Unix()
	for _, token := range tokenList {
		tokenInfo, err := s.fetchTokenInfo(token)
		if err != nil || tokenInfo == nil {
			continue
		}
		remain := tokenInfo.ExpireTime - now
		if remain > 0 && remain < min {
			min = remain
		}
	}
	if min <= 0 {
		min = s.config.Timeout
	}
	return min
}

func (s *SaToken) cleanTokenList(tokenList []string) []string {
	cleaned := make([]string, 0)
	for _, token := range tokenList {
		if token == "" {
			continue
		}
		if cache.Has(tokenKey(token)) {
			cleaned = append(cleaned, token)
		}
	}
	return cleaned
}

func (s *SaToken) removeTokenFromList(tokenList []string, token string) []string {
	result := make([]string, 0)
	for _, t := range tokenList {
		if t != token {
			result = append(result, t)
		}
	}
	return result
}

func (s *SaToken) acquireLock(lockKey string, ttl int, waitMs int) bool {
	if s.isRedisDriver {
		return s.acquireRedisLock(lockKey, ttl, waitMs)
	}
	return s.acquireMemoryLock(lockKey, waitMs)
}

func (s *SaToken) acquireRedisLock(lockKey string, ttl int, waitMs int) bool {
	start := time.Now().UnixMilli()
	for time.Now().UnixMilli()-start < int64(waitMs) {
		ok, _ := cache.SetNX(lockKey, "1", time.Duration(ttl)*time.Second)
		if ok {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func (s *SaToken) acquireMemoryLock(lockKey string, waitMs int) bool {
	start := time.Now().UnixMilli()
	for time.Now().UnixMilli()-start < int64(waitMs) {
		_, loaded := s.memoryLocks.LoadOrStore(lockKey, struct{}{})
		if !loaded {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func (s *SaToken) releaseLock(lockKey string) {
	if s.isRedisDriver {
		cache.Delete(lockKey)
	} else {
		s.memoryLocks.Delete(lockKey)
	}
}

func (s *SaToken) fetchTokenInfo(token string) (*TokenInfo, error) {
	if !s.validateTokenFormat(token) {
		return nil, ErrTokenFormat
	}

	val, err := cache.Get(tokenKey(token))
	if err != nil {
		return nil, ErrTokenInvalid
	}

	var info TokenInfo
	if err := json.Unmarshal([]byte(val), &info); err != nil {
		return nil, ErrTokenInvalid
	}

	return &info, nil
}

func (s *SaToken) renewIfNeeded(token string, tokenInfo *TokenInfo) {
	if !s.config.AutoRenew {
		return
	}

	now := time.Now().Unix()
	remaining := tokenInfo.ExpireTime - now

	if remaining >= int64(float64(s.config.Timeout)*s.config.RenewThreshold) {
		return
	}

	tokenInfo.ExpireTime = now + s.config.Timeout
	tokenInfo.CreateTime = now
	data, _ := json.Marshal(tokenInfo)
	cache.Set(tokenKey(token), string(data), time.Duration(s.config.Timeout)*time.Second)

	loginIDKey := loginIDKey(tokenInfo.LoginID)
	tokenList, err := s.getTokenList(tokenInfo.LoginID)
	if err == nil && len(tokenList) > 0 {
		ttl := s.getMinRemainingTime(tokenList)
		listData, _ := json.Marshal(tokenList)
		cache.Set(loginIDKey, string(listData), time.Duration(ttl)*time.Second)
	}

	logger.Debug("[SaToken] Token renewed: %s", token[:8])
}

func (s *SaToken) getTokenList(loginID int64) ([]string, error) {
	val, err := cache.Get(loginIDKey(loginID))
	if err != nil {
		return nil, err
	}
	if val == "" {
		return []string{}, nil
	}

	var list []string
	if err := json.Unmarshal([]byte(val), &list); err != nil {
		return []string{val}, nil
	}
	return list, nil
}

func Login(loginID int64, extra ...map[string]interface{}) (string, error) {
	s := getInstance()
	timeout := s.config.Timeout
	maxCount := s.resolveMaxLoginCount()

	token := s.createToken()

	tk := tokenKey(token)
	lik := loginIDKey(loginID)
	lk := lockKey("login:" + fmt.Sprintf("%d", loginID))

	locked := s.acquireLock(lk, 5, 500)
	if !locked {
		return "", errors.New("获取锁失败，请稍后重试")
	}
	defer s.releaseLock(lk)

	tokenList, err := s.getTokenList(loginID)
	if err != nil {
		tokenList = []string{}
	}
	tokenList = s.cleanTokenList(tokenList)

	tokenList = append(tokenList, token)

	for len(tokenList) > maxCount {
		oldest := tokenList[0]
		tokenList = tokenList[1:]
		cache.Delete(tokenKey(oldest))
	}

	ttl := s.getMinRemainingTime(tokenList)
	listData, _ := json.Marshal(tokenList)
	cache.Set(lik, string(listData), time.Duration(ttl)*time.Second)

	now := time.Now().Unix()
	tokenInfo := TokenInfo{
		LoginID:    loginID,
		CreateTime: now,
		ExpireTime: now + timeout,
	}
	if len(extra) > 0 && extra[0] != nil {
		tokenInfo.Extra = extra[0]
	}

	data, _ := json.Marshal(tokenInfo)
	cache.Set(tk, string(data), time.Duration(timeout)*time.Second)

	logger.Info("[SaToken] Login: login_id=%d, token=%s", loginID, token[:8])
	return token, nil
}

func Logout(token ...string) bool {
	s := getInstance()

	var t string
	if len(token) > 0 {
		t = token[0]
	} else {
		t = s.getTokenFromRequest()
	}

	if t == "" {
		return false
	}

	tokenInfo, err := s.fetchTokenInfo(t)
	if err != nil || tokenInfo == nil {
		return false
	}

	loginID := tokenInfo.LoginID
	tk := tokenKey(t)
	lik := loginIDKey(loginID)
	lk := lockKey("login:" + fmt.Sprintf("%d", loginID))

	locked := s.acquireLock(lk, 5, 200)
	if locked {
		defer s.releaseLock(lk)

		tokenList, _ := s.getTokenList(loginID)
		tokenList = s.removeTokenFromList(tokenList, t)

		if len(tokenList) == 0 {
			cache.Delete(lik)
		} else {
			ttl := s.getMinRemainingTime(tokenList)
			listData, _ := json.Marshal(tokenList)
			cache.Set(lik, string(listData), time.Duration(ttl)*time.Second)
		}
	}

	cache.Delete(tk)

	logger.Info("[SaToken] Logout: token=%s", t[:8])
	return true
}

func IsLogin(token ...string) bool {
	s := getInstance()

	var t string
	if len(token) > 0 {
		t = token[0]
	} else {
		t = s.getTokenFromRequest()
	}

	if t == "" {
		return false
	}

	tokenInfo, err := s.fetchTokenInfo(t)
	if err != nil || tokenInfo == nil {
		return false
	}

	if time.Now().Unix() > tokenInfo.ExpireTime {
		return false
	}

	s.renewIfNeeded(t, tokenInfo)

	return true
}

func CheckLogin(token ...string) error {
	s := getInstance()

	var t string
	if len(token) > 0 {
		t = token[0]
	} else {
		t = s.getTokenFromRequest()
	}

	if t == "" {
		return ErrNotLogin
	}

	tokenInfo, err := s.fetchTokenInfo(t)
	if err != nil {
		return err
	}

	if tokenInfo.LoginID == 0 {
		return ErrTokenInfo
	}

	if time.Now().Unix() > tokenInfo.ExpireTime {
		return ErrTokenExpired
	}

	s.renewIfNeeded(t, tokenInfo)

	return nil
}

func GetCurrentLoginId(token ...string) (int64, error) {
	s := getInstance()

	var t string
	if len(token) > 0 {
		t = token[0]
	} else {
		t = s.getTokenFromRequest()
	}

	if t == "" {
		return 0, ErrNotLogin
	}

	tokenInfo, err := s.fetchTokenInfo(t)
	if err != nil {
		return 0, err
	}

	if tokenInfo.LoginID == 0 {
		return 0, ErrTokenInfo
	}

	if time.Now().Unix() > tokenInfo.ExpireTime {
		return 0, ErrTokenExpired
	}

	s.renewIfNeeded(t, tokenInfo)

	return tokenInfo.LoginID, nil
}

func GetTokenInfo(token ...string) (*TokenInfo, error) {
	s := getInstance()

	var t string
	if len(token) > 0 {
		t = token[0]
	} else {
		t = s.getTokenFromRequest()
	}

	if t == "" {
		return nil, ErrNotLogin
	}

	tokenInfo, err := s.fetchTokenInfo(t)
	if err != nil {
		return nil, err
	}

	if time.Now().Unix() > tokenInfo.ExpireTime {
		return nil, ErrTokenExpired
	}

	s.renewIfNeeded(t, tokenInfo)

	return tokenInfo, nil
}

func GetExtra(token ...string) (map[string]interface{}, error) {
	tokenInfo, err := GetTokenInfo(token...)
	if err != nil {
		return nil, err
	}

	if tokenInfo.Extra == nil {
		return make(map[string]interface{}), nil
	}

	return tokenInfo.Extra, nil
}

func SetExtra(extra map[string]interface{}, token ...string) bool {
	s := getInstance()

	var t string
	if len(token) > 0 {
		t = token[0]
	} else {
		t = s.getTokenFromRequest()
	}

	if t == "" {
		return false
	}

	lk := lockKey("token:" + t)
	locked := s.acquireLock(lk, 3, 100)
	if locked {
		defer s.releaseLock(lk)
	}

	tokenInfo, err := s.fetchTokenInfo(t)
	if err != nil || tokenInfo == nil {
		return false
	}

	now := time.Now().Unix()
	if now > tokenInfo.ExpireTime {
		return false
	}

	tokenInfo.Extra = extra
	data, _ := json.Marshal(tokenInfo)
	remain := tokenInfo.ExpireTime - now
	cache.Set(tokenKey(t), string(data), time.Duration(remain)*time.Second)

	return true
}

func GetTokenExpireTime(token ...string) int64 {
	s := getInstance()

	var t string
	if len(token) > 0 {
		t = token[0]
	} else {
		t = s.getTokenFromRequest()
	}

	if t == "" {
		return 0
	}

	tokenInfo, err := s.fetchTokenInfo(t)
	if err != nil || tokenInfo == nil {
		return 0
	}

	return tokenInfo.ExpireTime
}

func GetTokenRemainingTime(token ...string) int64 {
	expire := GetTokenExpireTime(token...)
	remain := expire - time.Now().Unix()
	if remain < 0 {
		return 0
	}
	return remain
}

func Kickout(loginID int64) bool {
	s := getInstance()

	lik := loginIDKey(loginID)
	tokenList, err := s.getTokenList(loginID)
	if err != nil || len(tokenList) == 0 {
		return false
	}

	for _, token := range tokenList {
		cache.Delete(tokenKey(token))
	}

	cache.Delete(lik)

	logger.Info("[SaToken] Kickout: login_id=%d, tokens=%d", loginID, len(tokenList))
	return true
}

func KickoutByToken(token string) bool {
	s := getInstance()

	if token == "" {
		return false
	}

	tokenInfo, err := s.fetchTokenInfo(token)
	if err != nil || tokenInfo == nil {
		return false
	}

	return Logout(token)
}

func (s *SaToken) getTokenFromRequest() string {
	return ""
}
