package cache

import (
	"context"
	"fmt"
	"time"

	"lychee-go/internal/config"
	"lychee-go/internal/logger"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client
var ctx = context.Background()
var prefix string

// Init 初始化缓存
func Init() error {
	driver := config.GetString("cache.driver", "redis")

	if driver != "redis" {
		logger.Warn("Cache driver '%s' not fully implemented, skipping", driver)
		return nil
	}

	host := config.GetString("cache.host", "127.0.0.1")
	port := config.GetInt("cache.port", 6379)
	password := config.GetString("cache.password", "")
	database := config.GetInt("cache.database", 0)
	prefix = config.GetString("cache.prefix", "lychee_go_")

	RDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       database,
	})

	_, err := RDB.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect redis: %w", err)
	}

	logger.Info("Redis connected successfully (host=%s, db=%d)", host, database)
	return nil
}

func buildKey(key string) string {
	return prefix + key
}

// ======== 基础操作 ========

func Set(key string, value interface{}, expiration ...time.Duration) error {
	exp := time.Duration(config.GetInt("cache.default_ttl", 3600)) * time.Second
	if len(expiration) > 0 {
		exp = expiration[0]
	}
	return RDB.Set(ctx, buildKey(key), value, exp).Err()
}

func Get(key string) (string, error) {
	return RDB.Get(ctx, buildKey(key)).Result()
}

func Has(key string) bool {
	result, _ := RDB.Exists(ctx, buildKey(key)).Result()
	return result > 0
}

func Delete(key string) error {
	return RDB.Del(ctx, buildKey(key)).Err()
}

func Incr(key string) (int64, error) {
	return RDB.Incr(ctx, buildKey(key)).Result()
}

func Decr(key string) (int64, error) {
	return RDB.Decr(ctx, buildKey(key)).Result()
}

func TTL(key string) (time.Duration, error) {
	return RDB.TTL(ctx, buildKey(key)).Result()
}

func Expire(key string, expiration time.Duration) error {
	return RDB.Expire(ctx, buildKey(key), expiration).Err()
}

// ======== 分布式锁 ========

func SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	return RDB.SetNX(ctx, buildKey(key), value, expiration).Result()
}

// ======== Hash 操作 ========

func HSet(key, field string, value interface{}) error {
	return RDB.HSet(ctx, buildKey(key), field, value).Err()
}

func HGet(key, field string) (string, error) {
	return RDB.HGet(ctx, buildKey(key), field).Result()
}

func HGetAll(key string) (map[string]string, error) {
	return RDB.HGetAll(ctx, buildKey(key)).Result()
}

func HDel(key string, fields ...string) error {
	return RDB.HDel(ctx, buildKey(key), fields...).Err()
}

func HExists(key, field string) bool {
	result, _ := RDB.HExists(ctx, buildKey(key), field).Result()
	return result
}

// ======== List 操作 ========

func LPush(key string, values ...interface{}) error {
	return RDB.LPush(ctx, buildKey(key), values...).Err()
}

func RPop(key string) (string, error) {
	return RDB.RPop(ctx, buildKey(key)).Result()
}

func LLen(key string) (int64, error) {
	return RDB.LLen(ctx, buildKey(key)).Result()
}

func GetRedis() *redis.Client {
	return RDB
}
