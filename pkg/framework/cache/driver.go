package cache

import "time"

// Driver 缓存驱动接口
type Driver interface {
	// Set 设置缓存
	Set(key string, value interface{}, expiration time.Duration) error
	// Get 获取缓存
	Get(key string) (string, error)
	// Has 判断缓存是否存在
	Has(key string) bool
	// Delete 删除缓存
	Delete(key string) error
	// Incr 自增
	Incr(key string) (int64, error)
	// IncrBy 自增指定步长
	IncrBy(key string, delta int64) (int64, error)
	// Decr 自减
	Decr(key string) (int64, error)
	// DecrBy 自减指定步长
	DecrBy(key string, delta int64) (int64, error)
	// TTL 获取剩余过期时间
	TTL(key string) (time.Duration, error)
	// Expire 设置过期时间
	Expire(key string, expiration time.Duration) error
	// SetNX 设置值（如果不存在）
	SetNX(key string, value interface{}, expiration time.Duration) (bool, error)
	// HSet 设置 Hash 字段
	HSet(key, field string, value interface{}) error
	// HGet 获取 Hash 字段
	HGet(key, field string) (string, error)
	// HGetAll 获取 Hash 所有字段
	HGetAll(key string) (map[string]string, error)
	// HDel 删除 Hash 字段
	HDel(key string, fields ...string) error
	// HExists 判断 Hash 字段是否存在
	HExists(key, field string) bool
	// LPush 列表左侧推入
	LPush(key string, values ...interface{}) error
	// RPop 列表右侧弹出
	RPop(key string) (string, error)
	// LLen 获取列表长度
	LLen(key string) (int64, error)
	// ======== 标签操作 ========
	// SetWithTags 设置缓存并关联标签
	SetWithTags(key string, value interface{}, tags []string, expiration time.Duration) error
	// TagGet 获取标签下的所有缓存键
	TagGet(tagName string) ([]string, error)
	// TagFlush 删除标签下的所有缓存
	TagFlush(tagName string) error
	// TagClear 清除标签（不删除缓存）
	TagClear(tagName string) error
	// TagHas 判断标签下是否有缓存键
	TagHas(tagName string) bool
	// TagAdd 给标签添加缓存键
	TagAdd(tagName string, keys ...string) error
	// TagRemove 从标签中移除缓存键
	TagRemove(tagName string, keys ...string) error
}
