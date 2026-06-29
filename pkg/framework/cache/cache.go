package cache

import (
	"errors"
	"time"

	"github.com/watsonhaw5566/lychee-go/pkg/framework/config"
	"github.com/watsonhaw5566/lychee-go/pkg/framework/logger"

	"github.com/redis/go-redis/v9"
)

// ======== 管理器 ========

type Manager struct {
	defaultDriver string
	drivers       map[string]Driver
}

func NewManager() *Manager {
	return &Manager{
		drivers: make(map[string]Driver),
	}
}

func (m *Manager) Register(name string, driver Driver) {
	m.drivers[name] = driver
	if m.defaultDriver == "" {
		m.defaultDriver = name
	}
	logger.Info("[Cache] Driver registered: %s", name)
}

func (m *Manager) SetDefault(name string) {
	m.defaultDriver = name
}

func (m *Manager) Driver(name ...string) (Driver, error) {
	driverName := m.defaultDriver
	if len(name) > 0 {
		driverName = name[0]
	}

	driver, ok := m.drivers[driverName]
	if !ok {
		return nil, errors.New("cache driver not found: " + driverName)
	}
	return driver, nil
}

// ======== 全局实例 ========

var cacheManager *Manager
var RDB *redis.Client // 保留原有变量以兼容旧代码

// Init 初始化缓存
func Init() error {
	cacheManager = NewManager()

	driver := config.GetString("cache.driver", "memory")
	prefix := config.GetString("cache.prefix", "lychee_go_")

	// 内存驱动
	cacheManager.Register("memory", NewMemoryDriver(prefix))

	// 文件驱动
	fileRoot := config.GetString("cache.file.path", "runtime/cache")
	cacheManager.Register("file", NewFileDriver(fileRoot, prefix))

	// Redis 驱动
	if driver == "redis" || config.GetBool("cache.redis.enabled", false) {
		host := config.GetString("cache.host", "127.0.0.1")
		port := config.GetInt("cache.port", 6379)
		password := config.GetString("cache.password", "")
		database := config.GetInt("cache.database", 0)

		redisDriver, err := NewRedisDriver(host, port, password, database, prefix)
		if err != nil {
			logger.Warn("[Cache] Redis driver init failed: %v", err)
		} else {
			cacheManager.Register("redis", redisDriver)
			RDB = redisDriver.GetClient()
		}
	}

	cacheManager.SetDefault(driver)
	logger.Info("[Cache] Initialized (driver: %s)", driver)
	return nil
}

// ======== 对外 API ========

func GetDriver(name ...string) (Driver, error) {
	if cacheManager == nil {
		return nil, errors.New("cache not initialized, call cache.Init() first")
	}
	return cacheManager.Driver(name...)
}

func getDefaultDriver() (Driver, error) {
	return GetDriver()
}

// ======== 基础操作 ========

func Set(key string, value interface{}, expiration ...time.Duration) error {
	d, err := getDefaultDriver()
	if err != nil {
		return err
	}
	exp := time.Duration(config.GetInt("cache.default_ttl", 3600)) * time.Second
	if len(expiration) > 0 {
		exp = expiration[0]
	}
	return d.Set(key, value, exp)
}

func Get(key string) (string, error) {
	d, err := getDefaultDriver()
	if err != nil {
		return "", err
	}
	return d.Get(key)
}

func Has(key string) bool {
	d, err := getDefaultDriver()
	if err != nil {
		return false
	}
	return d.Has(key)
}

func Delete(key string) error {
	d, err := getDefaultDriver()
	if err != nil {
		return err
	}
	return d.Delete(key)
}

func Incr(key string) (int64, error) {
	d, err := getDefaultDriver()
	if err != nil {
		return 0, err
	}
	return d.Incr(key)
}

func IncrBy(key string, delta int64) (int64, error) {
	d, err := getDefaultDriver()
	if err != nil {
		return 0, err
	}
	return d.IncrBy(key, delta)
}

func Decr(key string) (int64, error) {
	d, err := getDefaultDriver()
	if err != nil {
		return 0, err
	}
	return d.Decr(key)
}

func DecrBy(key string, delta int64) (int64, error) {
	d, err := getDefaultDriver()
	if err != nil {
		return 0, err
	}
	return d.DecrBy(key, delta)
}

func TTL(key string) (time.Duration, error) {
	d, err := getDefaultDriver()
	if err != nil {
		return 0, err
	}
	return d.TTL(key)
}

func Expire(key string, expiration time.Duration) error {
	d, err := getDefaultDriver()
	if err != nil {
		return err
	}
	return d.Expire(key, expiration)
}

// ======== 分布式锁 ========

func SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	d, err := getDefaultDriver()
	if err != nil {
		return false, err
	}
	return d.SetNX(key, value, expiration)
}

// ======== Hash 操作 ========

func HSet(key, field string, value interface{}) error {
	d, err := getDefaultDriver()
	if err != nil {
		return err
	}
	return d.HSet(key, field, value)
}

func HGet(key, field string) (string, error) {
	d, err := getDefaultDriver()
	if err != nil {
		return "", err
	}
	return d.HGet(key, field)
}

func HGetAll(key string) (map[string]string, error) {
	d, err := getDefaultDriver()
	if err != nil {
		return nil, err
	}
	return d.HGetAll(key)
}

func HDel(key string, fields ...string) error {
	d, err := getDefaultDriver()
	if err != nil {
		return err
	}
	return d.HDel(key, fields...)
}

func HExists(key, field string) bool {
	d, err := getDefaultDriver()
	if err != nil {
		return false
	}
	return d.HExists(key, field)
}

// ======== List 操作 ========

func LPush(key string, values ...interface{}) error {
	d, err := getDefaultDriver()
	if err != nil {
		return err
	}
	return d.LPush(key, values...)
}

func RPop(key string) (string, error) {
	d, err := getDefaultDriver()
	if err != nil {
		return "", err
	}
	return d.RPop(key)
}

func LLen(key string) (int64, error) {
	d, err := getDefaultDriver()
	if err != nil {
		return 0, err
	}
	return d.LLen(key)
}

func GetRedis() *redis.Client {
	return RDB
}

// ======== 标签操作 ========

func SetWithTags(key string, value interface{}, tags []string, expiration ...time.Duration) error {
	d, err := getDefaultDriver()
	if err != nil {
		return err
	}
	exp := time.Duration(config.GetInt("cache.default_ttl", 3600)) * time.Second
	if len(expiration) > 0 {
		exp = expiration[0]
	}
	return d.SetWithTags(key, value, tags, exp)
}

// Tag 获取标签操作器
func Tag(tagName string) *TagHandler {
	return &TagHandler{tagName: tagName}
}

// TagHandler 标签操作器
type TagHandler struct {
	tagName string
}

func (t *TagHandler) Get() ([]string, error) {
	d, err := getDefaultDriver()
	if err != nil {
		return nil, err
	}
	return d.TagGet(t.tagName)
}

func (t *TagHandler) Flush() error {
	d, err := getDefaultDriver()
	if err != nil {
		return err
	}
	return d.TagFlush(t.tagName)
}

func (t *TagHandler) Clear() error {
	d, err := getDefaultDriver()
	if err != nil {
		return err
	}
	return d.TagClear(t.tagName)
}

func (t *TagHandler) Has() bool {
	d, err := getDefaultDriver()
	if err != nil {
		return false
	}
	return d.TagHas(t.tagName)
}

func (t *TagHandler) Add(keys ...string) error {
	d, err := getDefaultDriver()
	if err != nil {
		return err
	}
	return d.TagAdd(t.tagName, keys...)
}

func (t *TagHandler) Remove(keys ...string) error {
	d, err := getDefaultDriver()
	if err != nil {
		return err
	}
	return d.TagRemove(t.tagName, keys...)
}

func FlushTags(tags ...string) error {
	for _, tag := range tags {
		if err := Tag(tag).Flush(); err != nil {
			return err
		}
	}
	return nil
}
