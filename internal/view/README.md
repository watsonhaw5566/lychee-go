# view · 模板引擎模块

基于 Go Template 的 HTML 模板渲染模块，提供模板缓存、布局支持、自定义函数等功能。

## 功能特性

- ✅ Go Template 原生支持
- ✅ 模板缓存（开发/生产模式）
- ✅ 布局模板支持
- ✅ 内置常用模板函数
- ✅ 自定义函数注册
- ✅ 模板嵌套 include

## 配置

在 `config/config.yml` 中配置：

```yaml
view:
  path: resources/views    # 模板文件目录
  cache: false             # 是否启用模板缓存（生产环境建议开启）
  extension: .html         # 模板文件扩展名
  layout: layouts/main     # 全局布局模板（可选）
  content_key: content     # 布局模板中内容区块的变量名
```

## 使用示例

### 基本渲染

```go
package controller

import (
    "lychee-go/internal/view"
    "net/http"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
    data := map[string]interface{}{
        "Title": "首页",
        "Items": []string{"Item 1", "Item 2", "Item 3"},
    }
    view.Render(w, "index", data)
}
```

### 使用布局模板

```go
view.SetConfig(&view.Config{
    Layout: "layouts/main",
})

view.Render(w, "index", data)
```

### 注册自定义函数

```go
view.AddFunc("hello", func(name string) string {
    return "Hello, " + name + "!"
})

// 在模板中使用
// {{ hello "World" }}
```

### 模板嵌套

```html
<!-- views/header.html -->
<header>{{ .Title }}</header>

<!-- views/index.html -->
{{ include "header" . }}
<div>{{ .Content }}</div>
```

## 内置模板函数

### 字符串操作

| 函数 | 说明 | 示例 |
|------|------|------|
| `upper` | 转大写 | `{{ upper "hello" }}` → `HELLO` |
| `lower` | 转小写 | `{{ lower "HELLO" }}` → `hello` |
| `title` | 首字母大写 | `{{ title "hello world" }}` → `Hello World` |
| `trim` | 去除首尾空格 | `{{ trim "  hello  " }}` → `hello` |
| `replace` | 替换字符串 | `{{ replace "hello world" "world" "go" }}` |
| `split` | 分割字符串 | `{{ split "a,b,c" "," }}` → `["a", "b", "c"]` |
| `join` | 拼接字符串 | `{{ join .Items ", " }}` |
| `substr` | 截取子串 | `{{ substr "hello" 1 3 }}` → `ell` |

### 数字操作

| 函数 | 说明 | 示例 |
|------|------|------|
| `add` | 加法 | `{{ add 1 2 }}` → `3` |
| `sub` | 减法 | `{{ sub 5 3 }}` → `2` |
| `mul` | 乘法 | `{{ mul 2 3 }}` → `6` |
| `div` | 除法 | `{{ div 6 2 }}` → `3` |
| `mod` | 取模 | `{{ mod 7 3 }}` → `1` |

### 条件判断

| 函数 | 说明 | 示例 |
|------|------|------|
| `eq` | 相等 | `{{ if eq .A .B }}yes{{ end }}` |
| `ne` | 不相等 | `{{ if ne .A .B }}yes{{ end }}` |
| `gt` | 大于 | `{{ if gt .Count 10 }}yes{{ end }}` |
| `lt` | 小于 | `{{ if lt .Count 10 }}yes{{ end }}` |
| `ge` | 大于等于 | `{{ if ge .Count 10 }}yes{{ end }}` |
| `le` | 小于等于 | `{{ if le .Count 10 }}yes{{ end }}` |
| `and` | 逻辑与 | `{{ if and .A .B }}yes{{ end }}` |
| `or` | 逻辑或 | `{{ if or .A .B }}yes{{ end }}` |
| `not` | 逻辑非 | `{{ if not .Flag }}yes{{ end }}` |

### 格式化

| 函数 | 说明 | 示例 |
|------|------|------|
| `printf` | 格式化输出 | `{{ printf "Hello %s" .Name }}` |
| `html` | 输出 HTML | `{{ html "<b>bold</b>" }}` |
| `urlquery` | URL 编码 | `{{ urlquery .Query }}` |

### 模板嵌套

| 函数 | 说明 | 示例 |
|------|------|------|
| `include` | 包含子模板 | `{{ include "header" . }}` |

## 模板文件结构

```
resources/
└── views/
    ├── index.html          # 首页模板
    ├── users/
    │   ├── index.html      # 用户列表
    │   └── show.html       # 用户详情
    └── layouts/
        └── main.html       # 布局模板
```

## API 参考

### 渲染方法

- `Render(w io.Writer, name string, data interface{}) error` - 渲染模板到 Writer
- `RenderString(name string, data interface{}) (string, error)` - 渲染模板返回字符串
- `RenderWithLayout(w io.Writer, templateName, layoutName string, data interface{}) error` - 使用指定布局渲染

### 配置方法

- `Init()` - 初始化模板引擎
- `SetConfig(cfg *Config)` - 设置配置
- `GetConfig() *Config` - 获取当前配置

### 函数注册

- `AddFunc(name string, fn interface{})` - 注册单个函数
- `AddFuncs(funcs template.FuncMap)` - 批量注册函数

### 缓存管理

- `ClearCache()` - 清除模板缓存
- `GetCacheSize() int` - 获取缓存数量

### 模板管理

- `ListTemplates() ([]string, error)` - 获取所有模板列表
- `TemplateExists(name string) bool` - 检查模板是否存在