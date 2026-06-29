package validation

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/watsonhaw5566/lychee-go/pkg/framework/logger"
)

// ======== 自定义验证规则 ========

// ValidatorFunc 验证函数签名
type ValidatorFunc func(value interface{}, args ...string) bool

var (
	validators = map[string]ValidatorFunc{}
	rulesMu    sync.RWMutex
)

// RegisterValidator 注册自定义验证规则
func RegisterValidator(name string, fn ValidatorFunc) {
	rulesMu.Lock()
	defer rulesMu.Unlock()
	validators[name] = fn
	logger.Info("[Validation] Rule registered: %s", name)
}

// ======== 内置验证规则（类似 ThinkPHP 的验证规则） ========

func initBuiltin() {
	// 必填
	RegisterValidator("required", func(value interface{}, args ...string) bool {
		if value == nil {
			return false
		}
		v := reflect.ValueOf(value)
		switch v.Kind() {
		case reflect.String:
			return strings.TrimSpace(v.String()) != ""
		case reflect.Slice, reflect.Map, reflect.Array:
			return v.Len() > 0
		case reflect.Ptr, reflect.Interface:
			return !v.IsNil()
		}
		return true
	})

	// 邮箱
	RegisterValidator("email", func(value interface{}, args ...string) bool {
		s := toString(value)
		if s == "" {
			return true // 空值放行，配合 required
		}
		match, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, s)
		return match
	})

	// 手机号（中国大陆）
	RegisterValidator("mobile", func(value interface{}, args ...string) bool {
		s := toString(value)
		if s == "" {
			return true
		}
		match, _ := regexp.MatchString(`^1[3-9]\d{9}$`, s)
		return match
	})

	// URL
	RegisterValidator("url", func(value interface{}, args ...string) bool {
		s := toString(value)
		if s == "" {
			return true
		}
		match, _ := regexp.MatchString(`^https?://[^\s]+$`, s)
		return match
	})

	// 身份证
	RegisterValidator("idcard", func(value interface{}, args ...string) bool {
		s := toString(value)
		if s == "" {
			return true
		}
		match, _ := regexp.MatchString(`^\d{17}[\dXx]$`, s)
		return match
	})

	// 最小长度
	RegisterValidator("min", func(value interface{}, args ...string) bool {
		s := toString(value)
		if len(args) == 0 {
			return true
		}
		min, _ := strconv.Atoi(args[0])
		return len(s) >= min
	})

	// 最大长度
	RegisterValidator("max", func(value interface{}, args ...string) bool {
		s := toString(value)
		if len(args) == 0 {
			return true
		}
		max, _ := strconv.Atoi(args[0])
		return len(s) <= max
	})

	// 长度范围
	RegisterValidator("length", func(value interface{}, args ...string) bool {
		s := toString(value)
		if len(args) < 2 {
			return true
		}
		minLen, _ := strconv.Atoi(args[0])
		maxLen, _ := strconv.Atoi(args[1])
		return len(s) >= minLen && len(s) <= maxLen
	})

	// 数字
	RegisterValidator("number", func(value interface{}, args ...string) bool {
		s := toString(value)
		if s == "" {
			return true
		}
		_, err := strconv.ParseFloat(s, 64)
		return err == nil
	})

	// 整数
	RegisterValidator("integer", func(value interface{}, args ...string) bool {
		s := toString(value)
		if s == "" {
			return true
		}
		_, err := strconv.ParseInt(s, 10, 64)
		return err == nil
	})

	// 正整数
	RegisterValidator("positive", func(value interface{}, args ...string) bool {
		n, ok := toInt64(value)
		if !ok {
			return false
		}
		return n > 0
	})

	// 数值范围
	RegisterValidator("between", func(value interface{}, args ...string) bool {
		n, ok := toFloat64(value)
		if !ok || len(args) < 2 {
			return false
		}
		min, _ := strconv.ParseFloat(args[0], 64)
		max, _ := strconv.ParseFloat(args[1], 64)
		return n >= min && n <= max
	})

	// 相等
	RegisterValidator("eq", func(value interface{}, args ...string) bool {
		if len(args) == 0 {
			return true
		}
		return toString(value) == args[0]
	})

	// 不相等
	RegisterValidator("neq", func(value interface{}, args ...string) bool {
		if len(args) == 0 {
			return true
		}
		return toString(value) != args[0]
	})

	// 在列表中
	RegisterValidator("in", func(value interface{}, args ...string) bool {
		s := toString(value)
		if s == "" {
			return true
		}
		for _, arg := range args {
			if s == arg {
				return true
			}
		}
		return false
	})

	// 不在列表中
	RegisterValidator("notin", func(value interface{}, args ...string) bool {
		s := toString(value)
		if s == "" {
			return true
		}
		for _, arg := range args {
			if s == arg {
				return false
			}
		}
		return true
	})

	// 自定义正则
	RegisterValidator("regex", func(value interface{}, args ...string) bool {
		s := toString(value)
		if s == "" || len(args) == 0 {
			return true
		}
		match, _ := regexp.MatchString(args[0], s)
		return match
	})

	// 中文
	RegisterValidator("chinese", func(value interface{}, args ...string) bool {
		s := toString(value)
		if s == "" {
			return true
		}
		match, _ := regexp.MatchString(`^[\p{Han}]+$`, s)
		return match
	})

	// 字母数字下划线
	RegisterValidator("alpha_num", func(value interface{}, args ...string) bool {
		s := toString(value)
		if s == "" {
			return true
		}
		match, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, s)
		return match
	})

	// 用户名（字母开头，字母数字下划线，4-20位）
	RegisterValidator("username", func(value interface{}, args ...string) bool {
		s := toString(value)
		if s == "" {
			return true
		}
		match, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_]{3,19}$`, s)
		return match
	})

	// 密码（字母+数字，6-20位）
	RegisterValidator("password", func(value interface{}, args ...string) bool {
		s := toString(value)
		if s == "" {
			return true
		}
		hasLetter, _ := regexp.MatchString(`[a-zA-Z]`, s)
		hasNum, _ := regexp.MatchString(`[0-9]`, s)
		return hasLetter && hasNum && len(s) >= 6 && len(s) <= 20
	})

	// IPv4 地址
	RegisterValidator("ipv4", func(value interface{}, args ...string) bool {
		s := toString(value)
		if s == "" {
			return true
		}
		match, _ := regexp.MatchString(`^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$`, s)
		return match
	})
}

// 包初始化时注册内置规则
func init() {
	initBuiltin()
}

// ======== 工具函数 ========

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", val)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", val)
	case float32, float64:
		return fmt.Sprintf("%v", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", val)
	}
}

func toInt64(v interface{}) (int64, bool) {
	switch val := v.(type) {
	case int:
		return int64(val), true
	case int64:
		return val, true
	case int32:
		return int64(val), true
	case string:
		n, err := strconv.ParseInt(val, 10, 64)
		return n, err == nil
	}
	return 0, false
}

func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case string:
		f, err := strconv.ParseFloat(val, 64)
		return f, err == nil
	}
	return 0, false
}

// ======== 验证器 ========

// Validator 验证器实例
type Validator struct {
	rules      map[string]string // 字段 → 规则（用 | 分隔多个规则，用 : 分隔参数）
	messages   map[string]string // 字段 → 错误消息
	fieldNames map[string]string // 字段 → 显示名称
}

// New 创建验证器
func New() *Validator {
	return &Validator{
		rules:      make(map[string]string),
		messages:   make(map[string]string),
		fieldNames: make(map[string]string),
	}
}

// WithRule 添加字段验证规则
// 规则格式: "required|email|min:6|in:1,2,3"
func (v *Validator) WithRule(field, rules string) *Validator {
	v.rules[field] = rules
	return v
}

// WithMessage 添加自定义错误消息
// 格式: "field.rule" → "自定义消息"
func (v *Validator) WithMessage(key, message string) *Validator {
	v.messages[key] = message
	return v
}

// WithName 设置字段的中文名称（用于错误消息）
func (v *Validator) WithName(field, name string) *Validator {
	v.fieldNames[field] = name
	return v
}

// SetRules 批量设置规则
func (v *Validator) SetRules(rules map[string]string) *Validator {
	for f, r := range rules {
		v.rules[f] = r
	}
	return v
}

// SetMessages 批量设置消息
func (v *Validator) SetMessages(msgs map[string]string) *Validator {
	for k, m := range msgs {
		v.messages[k] = m
	}
	return v
}

// ======== 验证方法 ========

// Validate 验证 data（map 结构）
func (v *Validator) Validate(data map[string]interface{}) (bool, map[string]string) {
	errMap := make(map[string]string)

	for field, rules := range v.rules {
		value := data[field]
		fieldName := field
		if n, ok := v.fieldNames[field]; ok {
			fieldName = n
		}

		// 分割规则：按 | 分割
		ruleList := strings.Split(rules, "|")
		for _, rule := range ruleList {
			rule = strings.TrimSpace(rule)
			if rule == "" {
				continue
			}

			// 解析规则名和参数
			ruleParts := strings.SplitN(rule, ":", 2)
			ruleName := ruleParts[0]
			var ruleArgs []string
			if len(ruleParts) == 2 {
				ruleArgs = strings.Split(ruleParts[1], ",")
			}

			// required 不允许空值
			if ruleName != "required" && isEmpty(value) {
				continue
			}

			// 查找验证函数
			rulesMu.RLock()
			fn, ok := validators[ruleName]
			rulesMu.RUnlock()

			if !ok {
				errMap[field] = fmt.Sprintf("未知的验证规则: %s", ruleName)
				break
			}

			// 执行验证
			if !fn(value, ruleArgs...) {
				msgKey := field + "." + ruleName
				if msg, exists := v.messages[msgKey]; exists {
					errMap[field] = msg
				} else {
					errMap[field] = fmt.Sprintf("%s 格式不正确", fieldName)
				}
				break
			}
		}
	}

	return len(errMap) == 0, errMap
}

// ValidateStruct 验证结构体（通过反射读取字段）
func (v *Validator) ValidateStruct(data interface{}) (bool, map[string]string) {
	dataMap := make(map[string]interface{})

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return false, map[string]string{"_": "data must be a struct or map"}
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		fieldName := typ.Field(i).Name
		dataMap[strings.ToLower(fieldName)] = val.Field(i).Interface()
	}

	return v.Validate(dataMap)
}

// ======== 便捷函数 ========

// Check 单次验证一个值
func Check(value interface{}, rules string) (bool, error) {
	v := New()
	v.WithRule("field", rules)
	ok, errs := v.Validate(map[string]interface{}{"field": value})
	if !ok {
		return false, errors.New(errs["field"])
	}
	return true, nil
}

// IsEmail 判断是否邮箱
func IsEmail(email string) bool {
	ok, _ := Check(email, "required|email")
	return ok
}

// IsMobile 判断是否手机号
func IsMobile(mobile string) bool {
	ok, _ := Check(mobile, "required|mobile")
	return ok
}

// IsURL 判断是否 URL
func IsURL(url string) bool {
	ok, _ := Check(url, "required|url")
	return ok
}

// IsIDCard 判断是否身份证
func IsIDCard(idcard string) bool {
	ok, _ := Check(idcard, "required|idcard")
	return ok
}

// 辅助：判断空值
func isEmpty(v interface{}) bool {
	if v == nil {
		return true
	}
	s := toString(v)
	return s == ""
}

// ======== 错误信息辅助 ========

// FirstError 获取第一个错误
func FirstError(errors map[string]string) string {
	for _, v := range errors {
		return v
	}
	return ""
}

// AllErrors 获取所有错误（逗号拼接）
func AllErrors(errors map[string]string) string {
	var msgs []string
	for _, v := range errors {
		msgs = append(msgs, v)
	}
	return strings.Join(msgs, "; ")
}

// GetErrors 获取所有错误（map）
func GetErrors(errors map[string]string) []string {
	var msgs []string
	for _, v := range errors {
		msgs = append(msgs, v)
	}
	return msgs
}
