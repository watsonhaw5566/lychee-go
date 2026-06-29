package cache

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// FileDriver 文件缓存驱动
type FileDriver struct {
	rootDir   string
	prefix    string
	tagPrefix string
	mu        sync.RWMutex
	tagMu     sync.RWMutex
}

// fileItem 文件缓存项
type fileItem struct {
	Value      string    `json:"value"`
	ExpireTime time.Time `json:"expire_time"`
}

// tagItem 标签项
type tagItem struct {
	Keys []string `json:"keys"`
}

// NewFileDriver 创建文件驱动
func NewFileDriver(rootDir, prefix string) *FileDriver {
	return &FileDriver{
		rootDir:   strings.TrimRight(rootDir, "/"),
		prefix:    prefix,
		tagPrefix: prefix + "tag_",
	}
}

func (d *FileDriver) buildKey(key string) string {
	return d.prefix + key
}

func (d *FileDriver) buildTagKey(tagName string) string {
	return d.tagPrefix + tagName
}

func (d *FileDriver) getFilePath(key string) string {
	key = d.buildKey(key)
	hash := md5.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])
	return filepath.Join(d.rootDir, hashStr[:2], hashStr[2:4], hashStr[4:])
}

func (d *FileDriver) getTagFilePath(tagName string) string {
	tagKey := d.buildTagKey(tagName)
	hash := md5.Sum([]byte(tagKey))
	hashStr := hex.EncodeToString(hash[:])
	return filepath.Join(d.rootDir, "tags", hashStr[:2], hashStr[2:4], hashStr[4:])
}

func (d *FileDriver) Set(key string, value interface{}, expiration time.Duration) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	var val string
	switch v := value.(type) {
	case string:
		val = v
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		val = string(data)
	}

	item := &fileItem{
		Value:      val,
		ExpireTime: time.Now().Add(expiration),
	}

	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	filePath := d.getFilePath(key)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

func (d *FileDriver) Get(key string) (string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	filePath := d.getFilePath(key)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", errors.New("cache not found")
		}
		return "", err
	}

	var item fileItem
	if err := json.Unmarshal(data, &item); err != nil {
		return "", err
	}

	if !item.ExpireTime.IsZero() && time.Now().After(item.ExpireTime) {
		go os.Remove(filePath)
		return "", errors.New("cache not found")
	}

	return item.Value, nil
}

func (d *FileDriver) Has(key string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	filePath := d.getFilePath(key)
	_, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}

	var item fileItem
	if err := json.Unmarshal(data, &item); err != nil {
		return false
	}

	if !item.ExpireTime.IsZero() && time.Now().After(item.ExpireTime) {
		return false
	}

	return true
}

func (d *FileDriver) Delete(key string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	filePath := d.getFilePath(key)
	err := os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (d *FileDriver) Incr(key string) (int64, error) {
	return d.IncrBy(key, 1)
}

func (d *FileDriver) IncrBy(key string, delta int64) (int64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	filePath := d.getFilePath(key)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			item := &fileItem{
				Value:      "1",
				ExpireTime: time.Now().Add(time.Hour),
			}
			data, _ := json.Marshal(item)
			os.MkdirAll(filepath.Dir(filePath), 0755)
			os.WriteFile(filePath, data, 0644)
			return 1, nil
		}
		return 0, err
	}

	var item fileItem
	if err := json.Unmarshal(data, &item); err != nil {
		return 0, err
	}

	if !item.ExpireTime.IsZero() && time.Now().After(item.ExpireTime) {
		os.Remove(filePath)
		return 0, errors.New("cache not found")
	}

	var num int64
	if err := json.Unmarshal([]byte(item.Value), &num); err != nil {
		return 0, err
	}

	num += delta
	item.Value = fmt.Sprintf("%d", num)
	data, _ = json.Marshal(item)
	return num, os.WriteFile(filePath, data, 0644)
}

func (d *FileDriver) Decr(key string) (int64, error) {
	return d.DecrBy(key, 1)
}

func (d *FileDriver) DecrBy(key string, delta int64) (int64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	filePath := d.getFilePath(key)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, errors.New("cache not found")
		}
		return 0, err
	}

	var item fileItem
	if err := json.Unmarshal(data, &item); err != nil {
		return 0, err
	}

	if !item.ExpireTime.IsZero() && time.Now().After(item.ExpireTime) {
		os.Remove(filePath)
		return 0, errors.New("cache not found")
	}

	var num int64
	if err := json.Unmarshal([]byte(item.Value), &num); err != nil {
		return 0, err
	}

	num -= delta
	item.Value = fmt.Sprintf("%d", num)
	data, _ = json.Marshal(item)
	return num, os.WriteFile(filePath, data, 0644)
}

func (d *FileDriver) TTL(key string) (time.Duration, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	filePath := d.getFilePath(key)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return -1, errors.New("cache not found")
		}
		return -1, err
	}

	var item fileItem
	if err := json.Unmarshal(data, &item); err != nil {
		return -1, err
	}

	if time.Now().After(item.ExpireTime) {
		go os.Remove(filePath)
		return -1, errors.New("cache not found")
	}

	return time.Until(item.ExpireTime), nil
}

func (d *FileDriver) Expire(key string, expiration time.Duration) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	filePath := d.getFilePath(key)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("cache not found")
		}
		return err
	}

	var item fileItem
	if err := json.Unmarshal(data, &item); err != nil {
		return err
	}

	if time.Now().After(item.ExpireTime) {
		os.Remove(filePath)
		return errors.New("cache not found")
	}

	item.ExpireTime = time.Now().Add(expiration)
	data, _ = json.Marshal(item)
	return os.WriteFile(filePath, data, 0644)
}

func (d *FileDriver) SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	filePath := d.getFilePath(key)
	if _, err := os.Stat(filePath); err == nil {
		return false, nil
	}

	var val string
	switch v := value.(type) {
	case string:
		val = v
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return false, err
		}
		val = string(data)
	}

	item := &fileItem{
		Value:      val,
		ExpireTime: time.Now().Add(expiration),
	}

	data, err := json.Marshal(item)
	if err != nil {
		return false, err
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return false, err
	}

	return true, os.WriteFile(filePath, data, 0644)
}

// ======== Hash 操作 ========

func (d *FileDriver) HSet(key, field string, value interface{}) error {
	hashKey := key + ":" + field
	return d.Set(hashKey, value, 0)
}

func (d *FileDriver) HGet(key, field string) (string, error) {
	hashKey := key + ":" + field
	return d.Get(hashKey)
}

func (d *FileDriver) HGetAll(key string) (map[string]string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make(map[string]string)

	err := filepath.WalkDir(d.rootDir, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		var item fileItem
		if err := json.Unmarshal(data, &item); err != nil {
			return nil
		}

		if !item.ExpireTime.IsZero() && time.Now().After(item.ExpireTime) {
			return nil
		}

		return nil
	})

	return result, err
}

func (d *FileDriver) HDel(key string, fields ...string) error {
	for _, field := range fields {
		hashKey := key + ":" + field
		if err := d.Delete(hashKey); err != nil {
			return err
		}
	}
	return nil
}

func (d *FileDriver) HExists(key, field string) bool {
	hashKey := key + ":" + field
	return d.Has(hashKey)
}

// ======== List 操作 ========

func (d *FileDriver) LPush(key string, values ...interface{}) error {
	listKey := key + ":list"

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

	current, _ := d.Get(listKey)

	var list []string
	if current != "" {
		json.Unmarshal([]byte(current), &list)
	}

	list = append(items, list...)
	data, _ := json.Marshal(list)
	return d.Set(listKey, string(data), 0)
}

func (d *FileDriver) RPop(key string) (string, error) {
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

func (d *FileDriver) LLen(key string) (int64, error) {
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

func (d *FileDriver) SetWithTags(key string, value interface{}, tags []string, expiration time.Duration) error {
	if err := d.Set(key, value, expiration); err != nil {
		return err
	}

	for _, tag := range tags {
		d.TagAdd(tag, key)
	}

	return nil
}

func (d *FileDriver) TagGet(tagName string) ([]string, error) {
	d.tagMu.RLock()
	defer d.tagMu.RUnlock()

	filePath := d.getTagFilePath(tagName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var item tagItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, err
	}

	return item.Keys, nil
}

func (d *FileDriver) TagFlush(tagName string) error {
	d.tagMu.Lock()
	defer d.tagMu.Unlock()

	filePath := d.getTagFilePath(tagName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var item tagItem
	if err := json.Unmarshal(data, &item); err != nil {
		return err
	}

	for _, key := range item.Keys {
		d.Delete(key)
	}

	os.Remove(filePath)
	return nil
}

func (d *FileDriver) TagClear(tagName string) error {
	d.tagMu.Lock()
	defer d.tagMu.Unlock()

	filePath := d.getTagFilePath(tagName)
	err := os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (d *FileDriver) TagHas(tagName string) bool {
	d.tagMu.RLock()
	defer d.tagMu.RUnlock()

	filePath := d.getTagFilePath(tagName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}

	var item tagItem
	if err := json.Unmarshal(data, &item); err != nil {
		return false
	}

	return len(item.Keys) > 0
}

func (d *FileDriver) TagAdd(tagName string, keys ...string) error {
	d.tagMu.Lock()
	defer d.tagMu.Unlock()

	filePath := d.getTagFilePath(tagName)
	var item tagItem

	data, err := os.ReadFile(filePath)
	if err == nil {
		json.Unmarshal(data, &item)
	}

keyLoop:
	for _, key := range keys {
		for _, k := range item.Keys {
			if k == key {
				continue keyLoop
			}
		}
		item.Keys = append(item.Keys, key)
	}

	data, err = json.Marshal(item)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

func (d *FileDriver) TagRemove(tagName string, keys ...string) error {
	d.tagMu.Lock()
	defer d.tagMu.Unlock()

	filePath := d.getTagFilePath(tagName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var item tagItem
	if err := json.Unmarshal(data, &item); err != nil {
		return err
	}

	keySet := make(map[string]bool)
	for _, key := range keys {
		keySet[key] = true
	}

	var newKeys []string
	for _, k := range item.Keys {
		if !keySet[k] {
			newKeys = append(newKeys, k)
		}
	}
	item.Keys = newKeys

	data, err = json.Marshal(item)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}
