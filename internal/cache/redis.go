package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisDriver Redis 缓存驱动
type RedisDriver struct {
	client    *redis.Client
	prefix    string
	tagPrefix string
	ctx       context.Context
}

// NewRedisDriver 创建 Redis 驱动
func NewRedisDriver(host string, port int, password string, database int, prefix string) (*RedisDriver, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       database,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect redis: %w", err)
	}

	return &RedisDriver{
		client:    client,
		prefix:    prefix,
		tagPrefix: prefix + "tag_",
		ctx:       context.Background(),
	}, nil
}

func (d *RedisDriver) buildKey(key string) string {
	return d.prefix + key
}

func (d *RedisDriver) buildTagKey(tagName string) string {
	return d.tagPrefix + tagName
}

// ======== 基础操作 ========

func (d *RedisDriver) Set(key string, value interface{}, expiration time.Duration) error {
	return d.client.Set(d.ctx, d.buildKey(key), value, expiration).Err()
}

func (d *RedisDriver) Get(key string) (string, error) {
	return d.client.Get(d.ctx, d.buildKey(key)).Result()
}

func (d *RedisDriver) Has(key string) bool {
	result, _ := d.client.Exists(d.ctx, d.buildKey(key)).Result()
	return result > 0
}

func (d *RedisDriver) Delete(key string) error {
	return d.client.Del(d.ctx, d.buildKey(key)).Err()
}

func (d *RedisDriver) Incr(key string) (int64, error) {
	return d.client.Incr(d.ctx, d.buildKey(key)).Result()
}

func (d *RedisDriver) IncrBy(key string, delta int64) (int64, error) {
	return d.client.IncrBy(d.ctx, d.buildKey(key), delta).Result()
}

func (d *RedisDriver) Decr(key string) (int64, error) {
	return d.client.Decr(d.ctx, d.buildKey(key)).Result()
}

func (d *RedisDriver) DecrBy(key string, delta int64) (int64, error) {
	return d.client.DecrBy(d.ctx, d.buildKey(key), delta).Result()
}

func (d *RedisDriver) TTL(key string) (time.Duration, error) {
	return d.client.TTL(d.ctx, d.buildKey(key)).Result()
}

func (d *RedisDriver) Expire(key string, expiration time.Duration) error {
	return d.client.Expire(d.ctx, d.buildKey(key), expiration).Err()
}

func (d *RedisDriver) SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	return d.client.SetNX(d.ctx, d.buildKey(key), value, expiration).Result()
}

// ======== Hash 操作 ========

func (d *RedisDriver) HSet(key, field string, value interface{}) error {
	return d.client.HSet(d.ctx, d.buildKey(key), field, value).Err()
}

func (d *RedisDriver) HGet(key, field string) (string, error) {
	return d.client.HGet(d.ctx, d.buildKey(key), field).Result()
}

func (d *RedisDriver) HGetAll(key string) (map[string]string, error) {
	return d.client.HGetAll(d.ctx, d.buildKey(key)).Result()
}

func (d *RedisDriver) HDel(key string, fields ...string) error {
	return d.client.HDel(d.ctx, d.buildKey(key), fields...).Err()
}

func (d *RedisDriver) HExists(key, field string) bool {
	result, _ := d.client.HExists(d.ctx, d.buildKey(key), field).Result()
	return result
}

// ======== List 操作 ========

func (d *RedisDriver) LPush(key string, values ...interface{}) error {
	return d.client.LPush(d.ctx, d.buildKey(key), values...).Err()
}

func (d *RedisDriver) RPop(key string) (string, error) {
	return d.client.RPop(d.ctx, d.buildKey(key)).Result()
}

func (d *RedisDriver) LLen(key string) (int64, error) {
	return d.client.LLen(d.ctx, d.buildKey(key)).Result()
}

// ======== 标签操作 ========

func (d *RedisDriver) SetWithTags(key string, value interface{}, tags []string, expiration time.Duration) error {
	if err := d.Set(key, value, expiration); err != nil {
		return err
	}

	for _, tag := range tags {
		tagKey := d.buildTagKey(tag)
		d.client.SAdd(d.ctx, tagKey, key)
	}

	return nil
}

func (d *RedisDriver) TagGet(tagName string) ([]string, error) {
	tagKey := d.buildTagKey(tagName)
	return d.client.SMembers(d.ctx, tagKey).Result()
}

func (d *RedisDriver) TagFlush(tagName string) error {
	tagKey := d.buildTagKey(tagName)

	keys, err := d.client.SMembers(d.ctx, tagKey).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		var buildKeys []string
		for _, key := range keys {
			buildKeys = append(buildKeys, d.buildKey(key))
		}
		d.client.Del(d.ctx, buildKeys...)
	}

	d.client.Del(d.ctx, tagKey)
	return nil
}

func (d *RedisDriver) TagClear(tagName string) error {
	tagKey := d.buildTagKey(tagName)
	return d.client.Del(d.ctx, tagKey).Err()
}

func (d *RedisDriver) TagHas(tagName string) bool {
	tagKey := d.buildTagKey(tagName)
	count, _ := d.client.SCard(d.ctx, tagKey).Result()
	return count > 0
}

func (d *RedisDriver) TagAdd(tagName string, keys ...string) error {
	tagKey := d.buildTagKey(tagName)
	return d.client.SAdd(d.ctx, tagKey, keys).Err()
}

func (d *RedisDriver) TagRemove(tagName string, keys ...string) error {
	tagKey := d.buildTagKey(tagName)
	return d.client.SRem(d.ctx, tagKey, keys).Err()
}

// GetClient 获取原始 Redis 客户端
func (d *RedisDriver) GetClient() *redis.Client {
	return d.client
}
