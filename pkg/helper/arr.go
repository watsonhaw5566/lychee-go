package helper

import "reflect"

// InArray 判断值是否在数组中
func InArray(val interface{}, array interface{}) bool {
	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)
		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) {
				return true
			}
		}
	}
	return false
}

// ArrayUnique 数组去重
func ArrayUnique(arr []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(arr))
	for _, item := range arr {
		if _, ok := seen[item]; !ok {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

// ArrayIntersect 数组交集
func ArrayIntersect(a, b []string) []string {
	set := make(map[string]struct{})
	for _, v := range b {
		set[v] = struct{}{}
	}
	var result []string
	for _, v := range a {
		if _, ok := set[v]; ok {
			result = append(result, v)
		}
	}
	return result
}

// ArrayDiff 数组差集
func ArrayDiff(a, b []string) []string {
	set := make(map[string]struct{})
	for _, v := range b {
		set[v] = struct{}{}
	}
	var result []string
	for _, v := range a {
		if _, ok := set[v]; !ok {
			result = append(result, v)
		}
	}
	return result
}

// MapKeys 获取 Map 的所有 key
func MapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// MapValues 获取 Map 的所有 value
func MapValues(m map[string]interface{}) []interface{} {
	values := make([]interface{}, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// MapHasKey 判断 Map 是否包含指定 key
func MapHasKey(m map[string]interface{}, key string) bool {
	_, ok := m[key]
	return ok
}

// SliceChunk 按大小分割切片
func SliceChunk[T any](s []T, size int) [][]T {
	if size <= 0 {
		return [][]T{s}
	}
	var chunks [][]T
	for i := 0; i < len(s); i += size {
		end := i + size
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[i:end])
	}
	return chunks
}
