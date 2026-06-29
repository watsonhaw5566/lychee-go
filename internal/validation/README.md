# validation · 参数验证

类似 ThinkPHP 的字段验证器，支持 20+ 内置规则 + 自定义规则。

## 基本用法

```go
import "lychee-go/internal/validation"

v := validation.New()
v.WithRule("username", "required|min:3|max:20|alpha_num")
v.WithRule("email",    "required|email")
v.WithRule("password", "required|password")
v.WithRule("age",      "between:0,150")

ok, errors := v.Validate(map[string]interface{}{
    "username": "alice123",
    "email":    "alice@example.com",
    "password": "abc12345",
    "age":      25,
})

if !ok {
    // errors 是 map[string]string —— 字段 → 错误消息
    for field, msg := range errors {
        logger.Warn("%s: %s", field, msg)
    }
}
```

## 自定义错误消息 & 字段名

```go
v := validation.New()
v.WithRule("username", "required")
v.WithName("username", "用户名")           // 显示名
v.WithMessage("username.required", "请填写用户名")  // 自定义消息
```

## 批量设置

```go
v.SetRules(map[string]string{
    "email":    "required|email",
    "password": "required|password",
})

v.SetMessages(map[string]string{
    "email.required":    "邮箱不能为空",
    "email.email":       "邮箱格式不正确",
    "password.required": "密码不能为空",
})
```

## 验证结构体

```go
type RegisterReq struct {
    Username string
    Email    string
    Age      int
}

var req RegisterReq
c.ShouldBindJSON(&req)

// 通过反射读取字段
ok, errors := v.ValidateStruct(&req)
```

## 内置规则

| 规则 | 说明 | 示例 |
|------|------|------|
| `required` | 必填 | `"required"` |
| `email` | 邮箱格式 | `"email"` |
| `mobile` | 中国手机号 | `"mobile"` |
| `url` | URL 格式 | `"url"` |
| `idcard` | 身份证号 | `"idcard"` |
| `min:N` | 最小长度 | `"min:6"` |
| `max:N` | 最大长度 | `"max:20"` |
| `length:MIN,MAX` | 长度范围 | `"length:6,20"` |
| `number` | 数字 | `"number"` |
| `integer` | 整数 | `"integer"` |
| `positive` | 正整数 | `"positive"` |
| `between:MIN,MAX` | 数值范围 | `"between:0,150"` |
| `eq:VAL` | 相等 | `"eq:1"` |
| `neq:VAL` | 不相等 | `"neq:0"` |
| `in:A,B,C` | 在列表中 | `"in:1,2,3"` |
| `notin:A,B,C` | 不在列表中 | `"notin:0,999"` |
| `regex:PATTERN` | 自定义正则 | `"regex:^[a-z]+$"` |
| `chinese` | 纯中文 | `"chinese"` |
| `alpha_num` | 字母数字下划线 | `"alpha_num"` |
| `username` | 用户名规范 | `"username"` |
| `password` | 字母+数字，6-20 位 | `"password"` |
| `ipv4` | IPv4 地址 | `"ipv4"` |

## 自定义规则

```go
// 注册一条自定义规则
validation.RegisterValidator("is_even", func(value interface{}, args ...string) bool {
    n, _ := strconv.Atoi(fmt.Sprintf("%v", value))
    return n%2 == 0
})

// 使用
v.WithRule("count", "is_even")
```

## 便捷函数

```go
validation.IsEmail("test@example.com")     // true
validation.IsMobile("13800138000")          // true
validation.IsURL("https://example.com")     // true
validation.IsIDCard("110101199003071234")    // true

// 单次验证
ok, err := validation.Check(value, "required|email|min:5")
```

## 错误信息辅助

```go
firstMsg := validation.FirstError(errors)     // 获取第一条错误
allMsgs  := validation.AllErrors(errors)      // 逗号拼接的所有消息
list     := validation.GetErrors(errors)      // 字符串数组
```
