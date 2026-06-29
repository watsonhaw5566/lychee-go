package helper

import (
	"math/rand"
	"strings"
	"time"
)

// StrLen 获取字符串长度（支持中文）
func StrLen(str string) int {
	return len([]rune(str))
}

// SubStr 截取字符串
func SubStr(str string, start int, length ...int) string {
	runes := []rune(str)
	strLen := len(runes)

	if start < 0 {
		start = strLen + start
	}
	if start >= strLen {
		return ""
	}

	if len(length) > 0 {
		end := start + length[0]
		if end > strLen {
			end = strLen
		}
		if end < 0 {
			end = 0
		}
		return string(runes[start:end])
	}
	return string(runes[start:])
}

// StrUpper 转大写
func StrUpper(str string) string {
	return strings.ToUpper(str)
}

// StrLower 转小写
func StrLower(str string) string {
	return strings.ToLower(str)
}

// StrContains 是否包含子串
func StrContains(str, substr string) bool {
	return strings.Contains(str, substr)
}

// StrReplace 字符串替换
func StrReplace(search, replace, str string) string {
	return strings.ReplaceAll(str, search, replace)
}

// StrTrim 去除两端空白
func StrTrim(str string) string {
	return strings.TrimSpace(str)
}

// StrSplit 分割字符串
func StrSplit(str, sep string) []string {
	return strings.Split(str, sep)
}

// StrJoin 拼接字符串
func StrJoin(arr []string, sep string) string {
	return strings.Join(arr, sep)
}

// StrRepeat 重复字符串
func StrRepeat(str string, count int) string {
	return strings.Repeat(str, count)
}

// StrReverse 反转字符串
func StrReverse(str string) string {
	runes := []rune(str)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// RandomString 生成随机字符串
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}

// RandomNumber 生成随机数字字符串
func RandomNumber(length int) string {
	const charset = "0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}

// CamelToSnake 驼峰转下划线
func CamelToSnake(str string) string {
	var result []rune
	for i, r := range str {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

// SnakeToCamel 下划线转驼峰
func SnakeToCamel(str string) string {
	parts := strings.Split(str, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return strings.Join(parts, "")
}

// SnakeToLowerCamel 下划线转小驼峰
func SnakeToLowerCamel(str string) string {
	camel := SnakeToCamel(str)
	if len(camel) > 0 {
		return strings.ToLower(camel[:1]) + camel[1:]
	}
	return camel
}
