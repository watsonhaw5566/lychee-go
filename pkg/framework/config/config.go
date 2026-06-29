package config

import (
	"os"
	"regexp"
	"strings"

	"github.com/spf13/viper"
)

var v *viper.Viper

// 环境变量引用正则: ${VAR} 或 ${VAR:-default}
var envVarRegex = regexp.MustCompile(`\$\{([^}:-]+)(:-([^}]*))?\}`)

// Init 初始化配置（从单一 config.yml 加载）
func Init(configPath string) error {
	v = viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yml")
	v.AddConfigPath(configPath)
	v.AddConfigPath(".")
	v.AddConfigPath("../config")

	v.AutomaticEnv()
	v.SetEnvPrefix("LYCHEE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return err
	}

	// 解析配置中的环境变量引用
	resolveEnvVars()

	return nil
}

// resolveEnvVars 递归解析配置值中的环境变量引用
func resolveEnvVars() {
	for key := range v.AllSettings() {
		resolveKeyEnvVars(key)
	}
}

func resolveKeyEnvVars(key string) {
	val := v.Get(key)
	switch t := val.(type) {
	case string:
		resolved := expandEnvVars(t)
		if resolved != t {
			v.Set(key, resolved)
		}
	case map[string]interface{}:
		for k := range t {
			resolveKeyEnvVars(key + "." + k)
		}
	case []interface{}:
		for i, item := range t {
			if str, ok := item.(string); ok {
				resolved := expandEnvVars(str)
				if resolved != str {
					t[i] = resolved
					v.Set(key, t)
				}
			}
		}
	}
}

func expandEnvVars(value string) string {
	return envVarRegex.ReplaceAllStringFunc(value, func(match string) string {
		matches := envVarRegex.FindStringSubmatch(match)
		if len(matches) >= 2 {
			varName := matches[1]
			defaultValue := ""
			if len(matches) >= 4 {
				defaultValue = matches[3]
			}

			// 先从系统环境变量获取
			if val := os.Getenv(varName); val != "" {
				return val
			}
			return defaultValue
		}
		return match
	})
}

// Get 获取任意类型配置
func Get(key string, defaultValue ...interface{}) interface{} {
	if !v.IsSet(key) && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return v.Get(key)
}

// GetString 获取字符串
func GetString(key string, defaultValue ...string) string {
	if !v.IsSet(key) && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return v.GetString(key)
}

// GetInt 获取整数
func GetInt(key string, defaultValue ...int) int {
	if !v.IsSet(key) && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return v.GetInt(key)
}

// GetInt64 获取 64 位整数
func GetInt64(key string, defaultValue ...int64) int64 {
	if !v.IsSet(key) && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return v.GetInt64(key)
}

// GetBool 获取布尔
func GetBool(key string, defaultValue ...bool) bool {
	if !v.IsSet(key) && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return v.GetBool(key)
}

// GetFloat64 获取浮点数
func GetFloat64(key string, defaultValue ...float64) float64 {
	if !v.IsSet(key) && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return v.GetFloat64(key)
}

// GetStringSlice 获取字符串数组
func GetStringSlice(key string, defaultValue ...[]string) []string {
	if !v.IsSet(key) && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return v.GetStringSlice(key)
}

// AllSettings 获取所有配置
func AllSettings() map[string]interface{} {
	return v.AllSettings()
}

// IsSet 判断配置键是否实际存在（不是默认值，而是在 config.yml 中明确设置了）
func IsSet(key string) bool {
	if v == nil {
		return false
	}
	return v.IsSet(key)
}

// HasSection 判断某个配置块（section）是否实际配置了
// 例如：HasSection("database") 判断 config.yml 中是否存在 database 块
func HasSection(key string) bool {
	if v == nil {
		return false
	}
	return v.IsSet(key)
}

// IsConfigured 便捷方法：判断 key 对应的值是否实际存在且非空（字符串非空、数字非 0）
func IsConfigured(key string) bool {
	if v == nil || !v.IsSet(key) {
		return false
	}
	val := v.Get(key)
	switch t := val.(type) {
	case string:
		return strings.TrimSpace(t) != ""
	case int, int32, int64:
		return true // 数字只要存在就算配置了
	case float64:
		return true
	case bool:
		return true
	case nil:
		return false
	default:
		return true
	}
}

// GetViper 获取原生 Viper 实例
func GetViper() *viper.Viper {
	return v
}
