package cache

import (
	"encoding/json"
	"errors"
	"sync"
	"time"
)

// memoryItem 内存缓存项
type memoryItem struct {
	value      []byte
	expireTime time.Time
}

// MemoryDriver 内存缓存驱动
type MemoryDriver struct {
	data      sync.Map
	tagData   sync.Map
	prefix    string
	tagPrefix string
}

// NewMemoryDriver 创建内存驱动
func NewMemoryDriver(prefix string) *MemoryDriver {
	return &MemoryDriver{
		prefix:    prefix,
		tagPrefix: prefix + "tag_",
	}
}

func (d *MemoryDriver) buildKey(key string) string {
	return d.prefix + key
}

func (d *MemoryDriver) buildTagKey(tagName string) string {
	return d.tagPrefix + tagName
}

func (d *MemoryDriver) isExpired(item *memoryItem) bool {
	if item.expireTime.IsZero() {
		return false
	}
	return time.Now().After(item.expireTime)
}

func (d *MemoryDriver) Set(key string, value interface{}, expiration time.Duration) error {
	key = d.buildKey(key)

	var data []byte
	var err error

	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		data, err = json.Marshal(value)
		if err != nil {
			return err
		}
	}

	expireTime := time.Now().Add(expiration)
	d.data.Store(key, &memoryItem{
		value:      data,
		expireTime: expireTime,
	})
	return nil
}

func (d *MemoryDriver) Get(key string) (string, error) {
	key = d.buildKey(key)

	val, ok := d.data.Load(key)
	if !ok {
		return "", errors.New("cache not found")
	}

	item := val.(*memoryItem)
	if d.isExpired(item) {
		d.data.Delete(key)
		return "", errors.New("cache not found")
	}

	return string(item.value), nil
}

func (d *MemoryDriver) Has(key string) bool {
	key = d.buildKey(key)

	val, ok := d.data.Load(key)
	if !ok {
		return false
	}

	item := val.(*memoryItem)
	if d.isExpired(item) {
		d.data.Delete(key)
		return false
	}

	return true
}

func (d *MemoryDriver) Delete(key string) error {
	key = d.buildKey(key)
	d.data.Delete(key)
	return nil
}

func (d *MemoryDriver) Incr(key string) (int64, error) {
	return d.IncrBy(key, 1)
}

func (d *MemoryDriver) IncrBy(key string, delta int64) (int64, error) {
	key = d.buildKey(key)

	val, ok := d.data.Load(key)
	if !ok {
		d.data.Store(key, &memoryItem{
			value:      []byte("1"),
			expireTime: time.Now().Add(time.Hour),
		})
		return 1, nil
	}

	item := val.(*memoryItem)
	if d.isExpired(item) {
		d.data.Delete(key)
		return 0, errors.New("cache not found")
	}

	var num int64
	if err := json.Unmarshal(item.value, &num); err != nil {
		return 0, err
	}

	num += delta
	data, _ := json.Marshal(num)
	d.data.Store(key, &memoryItem{
		value:      data,
		expireTime: item.expireTime,
	})
	return num, nil
}

func (d *MemoryDriver) Decr(key string) (int64, error) {
	return d.DecrBy(key, 1)
}

func (d *MemoryDriver) DecrBy(key string, delta int64) (int64, error) {
	key = d.buildKey(key)

	val, ok := d.data.Load(key)
	if !ok {
		return 0, errors.New("cache not found")
	}

	item := val.(*memoryItem)
	if d.isExpired(item) {
		d.data.Delete(key)
		return 0, errors.New("cache not found")
	}

	var num int64
	if err := json.Unmarshal(item.value, &num); err != nil {
		return 0, err
	}

	num -= delta
	data, _ := json.Marshal(num)
	d.data.Store(key, &memoryItem{
		value:      data,
		expireTime: item.expireTime,
	})
	return num, nil
}

func (d *MemoryDriver) TTL(key string) (time.Duration, error) {
	key = d.buildKey(key)

	val, ok := d.data.Load(key)
	if !ok {
		return -1, errors.New("cache not found")
	}

	item := val.(*memoryItem)
	if d.isExpired(item) {
		d.data.Delete(key)
		return -1, errors.New("cache not found")
	}

	return time.Until(item.expireTime), nil
}

func (d *MemoryDriver) Expire(key string, expiration time.Duration) error {
	key = d.buildKey(key)

	val, ok := d.data.Load(key)
	if !ok {
		return errors.New("cache not found")
	}

	item := val.(*memoryItem)
	if d.isExpired(item) {
		d.data.Delete(key)
		return errors.New("cache not found")
	}

	item.expireTime = time.Now().Add(expiration)
	d.data.Store(key, item)
	return nil
}

func (d *MemoryDriver) SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	key = d.buildKey(key)

	var data []byte
	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		var err error
		data, err = json.Marshal(value)
		if err != nil {
			return false, err
		}
	}

	_, loaded := d.data.LoadOrStore(key, &memoryItem{
		value:      data,
		expireTime: time.Now().Add(expiration),
	})

	return !loaded, nil
}

// ======== Hash 操作 ========

func (d *MemoryDriver) HSet(key, field string, value interface{}) error {
	hashKey := d.buildKey(key) + ":" + field
	return d.Set(hashKey, value, 0)
}

func (d *MemoryDriver) HGet(key, field string) (string, error) {
	hashKey := d.buildKey(key) + ":" + field
	return d.Get(hashKey)
}

func (d *MemoryDriver) HGetAll(key string) (map[string]string, error) {
	key = d.buildKey(key)
	result := make(map[string]string)

	d.data.Range(func(k, v interface{}) bool {
		kStr := k.(string)
		if len(kStr) > len(key) && kStr[:len(key)] == key && len(kStr) > len(key)+1 && kStr[len(key)] == ':' {
			field := kStr[len(key)+1:]
			item := v.(*memoryItem)
			if !d.isExpired(item) {
				result[field] = string(item.value)
			}
		}
		return true
	})

	return result, nil
}

func (d *MemoryDriver) HDel(key string, fields ...string) error {
	key = d.buildKey(key)
	for _, field := range fields {
		hashKey := key + ":" + field
		d.data.Delete(hashKey)
	}
	return nil
}

func (d *MemoryDriver) HExists(key, field string) bool {
	hashKey := d.buildKey(key) + ":" + field
	return d.Has(hashKey)
}

// ======== List 操作 ========

func (d *MemoryDriver) LPush(key string, values ...interface{}) error {
	key = d.buildKey(key)

	var items []string
	for _, v := range values {
		var item string
		switch val := v.(type) {
		case string:
			item = val
		default:
			data, err := json.Marshal(v)
			if err != nil {
				return err
			}
			item = string(data)
		}
		items = append(items, item)
	}

	listKey := key + ":list"
	current, _ := d.Get(listKey)

	var list []string
	if current != "" {
		json.Unmarshal([]byte(current), &list)
	}

	list = append(items, list...)
	data, _ := json.Marshal(list)
	return d.Set(listKey, string(data), 0)
}

func (d *MemoryDriver) RPop(key string) (string, error) {
	key = d.buildKey(key)
	listKey := key + ":list"

	current, err := d.Get(listKey)
	if err != nil {
		return "", err
	}

	var list []string
	if err := json.Unmarshal([]byte(current), &list); err != nil {
		return "", err
	}

	if len(list) == 0 {
		return "", errors.New("list is empty")
	}

	result := list[len(list)-1]
	list = list[:len(list)-1]

	data, _ := json.Marshal(list)
	d.Set(listKey, string(data), 0)

	return result, nil
}

func (d *MemoryDriver) LLen(key string) (int64, error) {
	key = d.buildKey(key)
	listKey := key + ":list"

	current, err := d.Get(listKey)
	if err != nil {
		return 0, err
	}

	var list []string
	if err := json.Unmarshal([]byte(current), &list); err != nil {
		return 0, err
	}

	return int64(len(list)), nil
}

// ======== 标签操作 ========

func (d *MemoryDriver) SetWithTags(key string, value interface{}, tags []string, expiration time.Duration) error {
	if err := d.Set(key, value, expiration); err != nil {
		return err
	}

	for _, tag := range tags {
		tagKey := d.buildTagKey(tag)
		val, _ := d.tagData.LoadOrStore(tagKey, &sync.Map{})
		tagSet := val.(*sync.Map)
		tagSet.Store(key, struct{}{})
	}

	return nil
}

func (d *MemoryDriver) TagGet(tagName string) ([]string, error) {
	tagKey := d.buildTagKey(tagName)
	val, ok := d.tagData.Load(tagKey)
	if !ok {
		return []string{}, nil
	}

	tagSet := val.(*sync.Map)
	result := []string{}
	tagSet.Range(func(k, v interface{}) bool {
		result = append(result, k.(string))
		return true
	})
	return result, nil
}

func (d *MemoryDriver) TagFlush(tagName string) error {
	tagKey := d.buildTagKey(tagName)
	val, ok := d.tagData.Load(tagKey)
	if !ok {
		return nil
	}

	tagSet := val.(*sync.Map)
	tagSet.Range(func(k, v interface{}) bool {
		d.Delete(k.(string))
		return true
	})

	d.tagData.Delete(tagKey)
	return nil
}

func (d *MemoryDriver) TagClear(tagName string) error {
	tagKey := d.buildTagKey(tagName)
	d.tagData.Delete(tagKey)
	return nil
}

func (d *MemoryDriver) TagHas(tagName string) bool {
	tagKey := d.buildTagKey(tagName)
	val, ok := d.tagData.Load(tagKey)
	if !ok {
		return false
	}

	tagSet := val.(*sync.Map)
	has := false
	tagSet.Range(func(k, v interface{}) bool {
		has = true
		return false
	})
	return has
}

func (d *MemoryDriver) TagAdd(tagName string, keys ...string) error {
	tagKey := d.buildTagKey(tagName)
	val, _ := d.tagData.LoadOrStore(tagKey, &sync.Map{})
	tagSet := val.(*sync.Map)

	for _, key := range keys {
		tagSet.Store(key, struct{}{})
	}
	return nil
}

func (d *MemoryDriver) TagRemove(tagName string, keys ...string) error {
	tagKey := d.buildTagKey(tagName)
	val, ok := d.tagData.Load(tagKey)
	if !ok {
		return nil
	}

	tagSet := val.(*sync.Map)
	for _, key := range keys {
		tagSet.Delete(key)
	}
	return nil
}
