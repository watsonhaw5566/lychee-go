package env

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/watsonhaw5566/lychee-go/pkg/framework/logger"
)

var envCache = make(map[string]string)

// Load 加载 .env 文件
// 支持的格式：
// - KEY=VALUE
// - KEY="VALUE"
// - KEY='VALUE'
// - # 注释
// - ${VAR:-default} 环境变量引用，带默认值
func Load(paths ...string) error {
	if len(paths) == 0 {
		paths = []string{".env"}
	}

	for _, path := range paths {
		if err := loadFile(path); err != nil {
			logger.Warn("[Env] Load %s failed: %v", path, err)
		}
	}

	return nil
}

func loadFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open file failed: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 解析 KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			logger.Warn("[Env] Invalid line %d: %s", lineNum, line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// 移除引号
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}

		// 解析环境变量引用 ${VAR:-default}
		value = expandEnvVars(value)

		// 设置到系统环境变量和缓存
		os.Setenv(key, value)
		envCache[key] = value

		logger.Debug("[Env] Loaded: %s=%s", key, maskSecret(value))
	}

	return scanner.Err()
}

// 环境变量引用正则: ${VAR} 或 ${VAR:-default}
var envVarRegex = regexp.MustCompile(`\$\{([^}:-]+)(:-([^}]*))?\}`)

func expandEnvVars(value string) string {
	return envVarRegex.ReplaceAllStringFunc(value, func(match string) string {
		matches := envVarRegex.FindStringSubmatch(match)
		if len(matches) >= 2 {
			varName := matches[1]
			defaultValue := ""
			if len(matches) >= 4 {
				defaultValue = matches[3]
			}

			// 先从缓存找，再从系统环境变量找
			if val, ok := envCache[varName]; ok {
				return val
			}
			if val := os.Getenv(varName); val != "" {
				return val
			}
			return defaultValue
		}
		return match
	})
}

// Get 获取环境变量值
func Get(key string, defaultValue ...string) string {
	if val, ok := envCache[key]; ok {
		return val
	}
	if val := os.Getenv(key); val != "" {
		return val
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

// GetInt 获取环境变量的整数值
func GetInt(key string, defaultValue ...int) int {
	val := Get(key)
	if val == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}
	var result int
	fmt.Sscanf(val, "%d", &result)
	return result
}

// GetBool 获取环境变量的布尔值
func GetBool(key string, defaultValue ...bool) bool {
	val := strings.ToLower(Get(key))
	if val == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return false
	}
	return val == "true" || val == "1" || val == "yes"
}

// Set 设置环境变量
func Set(key, value string) {
	os.Setenv(key, value)
	envCache[key] = value
}

// All 获取所有已加载的环境变量
func All() map[string]string {
	result := make(map[string]string)
	for k, v := range envCache {
		result[k] = v
	}
	return result
}

// maskSecret 屏蔽敏感信息
func maskSecret(value string) string {
	if len(value) <= 8 {
		return "***"
	}
	return value[:4] + "***" + value[len(value)-4:]
}
