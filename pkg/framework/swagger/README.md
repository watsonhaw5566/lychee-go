# swagger · API 文档生成器

一个无第三方依赖的轻量 Swagger 文档生成器，你可以：

- 用代码注册端点（参数 / 响应示例）
- 访问 `http://localhost:8080/swagger` 在线浏览 Swagger UI
- 访问 `http://localhost:8080/swagger/doc.json` 获取 Swagger 2.0 JSON
- 在终端执行 `lychee-go doc` 打印所有已注册端点摘要

---

## 1. 初始化

`boot.Boot()` 已自动调用 `swagger.Init()`，并在 main.go 中注册了路由。
不需要你额外做什么。

`config.yml` 中的相关配置：

```yaml
swagger:
  enabled: true    # 生产环境可设 false 禁用
  path: "/swagger" # Swagger UI 和 JSON 的根路径
```

---

## 2. 便捷 API：注册一个端点

最常用的写法是 `swagger.GET` / `swagger.POST` / `swagger.PUT` / `swagger.DELETE`：

```go
package controller

import "lychee-go/internal/swagger"

func init() {
    // 注册一个 GET /api/users —— 列出用户
    swagger.GET("/api/users", "列出所有用户",
        []swagger.Parameter{
            swagger.P("page", "query", "页码", false, "integer"),
            swagger.P("size", "query", "每页数量", false, "integer"),
        },
        swagger.OK200(map[string]interface{}{
            "code":    0,
            "message": "ok",
            "data": []map[string]interface{}{
                {"id": 1, "username": "alice"},
                {"id": 2, "username": "bob"},
            },
        }),
    )

    // 注册一个 POST /api/users —— 创建用户
    swagger.POST("/api/users", "创建用户",
        []swagger.Parameter{
            swagger.P("username", "query", "用户名", true, "string"),
            swagger.P("email", "query", "邮箱", true, "string"),
        },
        swagger.OK200(map[string]interface{}{
            "code":    0,
            "message": "ok",
            "data":    map[string]interface{}{"id": 3},
        }),
    )
}
```

### `swagger.P(name, in, desc, required, type...)` 快速定义参数

- `name`：参数名
- `in`：来源 —— `"query" | "path" | "header" | "body" | "formData"`
- `desc`：描述（字符串）
- `required`：是否必填
- `type...`：可选的类型；第一个是 `"string" | "integer" | "number" | "boolean" | "array"`，第二个是 format

### `swagger.OK200(example)` 成功响应

传入一个 interface{} 示例即可，内部会生成 `200` 的响应描述。

---

## 3. 完整 API：`AddEndpoint`

如果 `GET / POST` 这类便捷方法不够用，直接用底层的 `AddEndpoint`：

```go
swagger.AddEndpoint("PATCH", "/api/users/:id", swagger.Endpoint{
    Tags:        []string{"users"},
    Summary:     "部分更新用户",
    Description: "只需传要更新的字段",
    Parameters: []swagger.Parameter{
        {Name: "id", In: "path", Required: true, Type: "integer", Description: "用户 ID"},
        {Name: "email", In: "formData", Required: false, Type: "string"},
    },
    Responses: map[string]swagger.Response{
        "200": swagger.R("成功", map[string]interface{}{"code": 0, "message": "ok"}),
        "404": swagger.Err404(),
    },
})
```

---

## 4. 手动访问 JSON / 文档

启动服务器后：

```
# Swagger UI
http://localhost:8080/swagger

# JSON（Swagger 2.0）
http://localhost:8080/swagger/doc.json

# 也可以用命令行看
lychee-go doc
```

---

## 5. 设计要点

1. **0 第三方依赖**：纯 Go + gin 实现，Swagger UI HTML 从 unpkg CDN 加载（不内置静态文件，保持仓库小）。
2. **接口式注册**：任何调用 `swagger.GET` 等函数的地方都能注册端点（推荐写在各自 controller 文件的 `init()` 里，代码即文档）。
3. **自动扫描兜底**：即使用户完全没写 swagger 注册，`swagger.ScanGinRoutes(r)` 也会自动把所有已注册 Gin 路由补成一条最简记录，保证 `GET /swagger` 有内容可看。
4. **生产环境可一键禁用**：改 `swagger.enabled: false` 即不注册 `/swagger` 路由。
