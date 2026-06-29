package swagger

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"sort"
	"sync"

	"github.com/watsonhaw5566/lychee-go/pkg/framework/config"
	"github.com/watsonhaw5566/lychee-go/pkg/framework/logger"

	"github.com/gin-gonic/gin"
)

// ============================================================
// 数据结构：Swagger 2.0 / OpenAPI 3.0 的简化实现
// ============================================================

// Info API 基本信息
type Info struct {
	Title          string   `json:"title"`
	Description    string   `json:"description,omitempty"`
	Version        string   `json:"version"`
	TermsOfService string   `json:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty"`
	License        *License `json:"license,omitempty"`
}

type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

type License struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// Parameter 参数定义
type Parameter struct {
	Name        string      `json:"name"`
	In          string      `json:"in"` // query, path, header, body, formData
	Description string      `json:"description,omitempty"`
	Required    bool        `json:"required,omitempty"`
	Type        string      `json:"type,omitempty"`   // string, integer, number, boolean, array
	Format      string      `json:"format,omitempty"` // int64, float, date, email...
	Default     interface{} `json:"default,omitempty"`
	Example     interface{} `json:"example,omitempty"`
}

// Response 响应定义
type Response struct {
	Description string                 `json:"description"`
	Example     interface{}            `json:"example,omitempty"`
	Schema      map[string]interface{} `json:"schema,omitempty"`
}

// Endpoint 一个具体的 API 端点
type Endpoint struct {
	Method      string              `json:"-"` // GET, POST, PUT, DELETE, PATCH
	Path        string              `json:"-"` // /api/users
	Tags        []string            `json:"tags,omitempty"`
	Summary     string              `json:"summary,omitempty"`
	Description string              `json:"description,omitempty"`
	OperationID string              `json:"operationId,omitempty"`
	Deprecated  bool                `json:"deprecated,omitempty"`
	Parameters  []Parameter         `json:"parameters,omitempty"`
	Responses   map[string]Response `json:"responses,omitempty"`
}

// Swagger 完整文档（Swagger 2.0 兼容格式）
type Swagger struct {
	Swagger  string                         `json:"swagger"`
	Info     Info                           `json:"info"`
	Host     string                         `json:"host,omitempty"`
	BasePath string                         `json:"basePath,omitempty"`
	Schemes  []string                       `json:"schemes,omitempty"`
	Consumes []string                       `json:"consumes,omitempty"`
	Produces []string                       `json:"produces,omitempty"`
	Paths    map[string]map[string]Endpoint `json:"paths"`
}

// ============================================================
// 全局状态
// ============================================================

var (
	doc      Swagger
	docMu    sync.RWMutex
	once     sync.Once
	hasBuild bool
)

// ============================================================
// 初始化
// ============================================================

// Init 初始化 Swagger 文档
// 在 main.go 中调用：swagger.Init()
func Init() {
	once.Do(func() {
		doc = Swagger{
			Swagger: "2.0",
			Info: Info{
				Title:       config.GetString("app.name", "Lychee-Go"),
				Description: "API Documentation for " + config.GetString("app.name", "Lychee-Go"),
				Version:     config.GetString("app.version", "1.0.0"),
				Contact: &Contact{
					Name:  "API Support",
					Email: "support@example.com",
				},
			},
			BasePath: "/",
			Schemes:  []string{"http", "https"},
			Consumes: []string{"application/json"},
			Produces: []string{"application/json"},
			Paths:    make(map[string]map[string]Endpoint),
		}
		hasBuild = false
		logger.Info("[Swagger] Initialized (%s v%s)",
			doc.Info.Title, doc.Info.Version)
	})
}

// SetTitle 设置文档标题
func SetTitle(title string) {
	docMu.Lock()
	defer docMu.Unlock()
	doc.Info.Title = title
}

// SetDescription 设置描述
func SetDescription(desc string) {
	docMu.Lock()
	defer docMu.Unlock()
	doc.Info.Description = desc
}

// SetVersion 设置版本
func SetVersion(version string) {
	docMu.Lock()
	defer docMu.Unlock()
	doc.Info.Version = version
}

// SetHost 设置主机名（如 api.example.com）
func SetHost(host string) {
	docMu.Lock()
	defer docMu.Unlock()
	doc.Host = host
}

// SetBasePath 设置基础路径
func SetBasePath(basePath string) {
	docMu.Lock()
	defer docMu.Unlock()
	doc.BasePath = basePath
}

// SetSchemes 设置协议
func SetSchemes(schemes ...string) {
	docMu.Lock()
	defer docMu.Unlock()
	doc.Schemes = schemes
}

// ============================================================
// 注册 API 端点
// ============================================================

// AddEndpoint 注册一个 API 端点
func AddEndpoint(method, urlPath string, endpoint Endpoint) {
	docMu.Lock()
	defer docMu.Unlock()

	// 规范化 method 为小写
	m := ""
	switch method {
	case http.MethodGet:
		m = "get"
	case http.MethodPost:
		m = "post"
	case http.MethodPut:
		m = "put"
	case http.MethodDelete:
		m = "delete"
	case http.MethodPatch:
		m = "patch"
	case http.MethodHead:
		m = "head"
	case http.MethodOptions:
		m = "options"
	default:
		m = "get"
	}

	endpoint.Method = m
	endpoint.Path = urlPath

	if doc.Paths == nil {
		doc.Paths = make(map[string]map[string]Endpoint)
	}
	if _, exists := doc.Paths[urlPath]; !exists {
		doc.Paths[urlPath] = make(map[string]Endpoint)
	}
	doc.Paths[urlPath][m] = endpoint
	hasBuild = false
}

// ============================================================
// 便捷注册方法
// ============================================================

// GET 注册一个 GET 端点
func GET(urlPath string, summary string, params []Parameter, responses map[string]Response) {
	AddEndpoint(http.MethodGet, urlPath, Endpoint{
		Summary:    summary,
		Parameters: params,
		Responses:  responses,
	})
}

// POST 注册一个 POST 端点
func POST(urlPath string, summary string, params []Parameter, responses map[string]Response) {
	AddEndpoint(http.MethodPost, urlPath, Endpoint{
		Summary:    summary,
		Parameters: params,
		Responses:  responses,
	})
}

// PUT 注册一个 PUT 端点
func PUT(urlPath string, summary string, params []Parameter, responses map[string]Response) {
	AddEndpoint(http.MethodPut, urlPath, Endpoint{
		Summary:    summary,
		Parameters: params,
		Responses:  responses,
	})
}

// DELETE 注册一个 DELETE 端点
func DELETE(urlPath string, summary string, params []Parameter, responses map[string]Response) {
	AddEndpoint(http.MethodDelete, urlPath, Endpoint{
		Summary:    summary,
		Parameters: params,
		Responses:  responses,
	})
}

// PATCH 注册一个 PATCH 端点
func PATCH(urlPath string, summary string, params []Parameter, responses map[string]Response) {
	AddEndpoint(http.MethodPatch, urlPath, Endpoint{
		Summary:    summary,
		Parameters: params,
		Responses:  responses,
	})
}

// ============================================================
// 参数 / 响应的便捷构造
// ============================================================

// P 快速创建一个 Parameter
//
//	Usage: swagger.P("id", "path", "用户 ID", true, "integer")
func P(name, in, desc string, required bool, typ ...string) Parameter {
	p := Parameter{Name: name, In: in, Description: desc, Required: required}
	if len(typ) > 0 {
		p.Type = typ[0]
	}
	if len(typ) > 1 {
		p.Format = typ[1]
	}
	return p
}

// R 快速创建一个 Response
func R(desc string, example ...interface{}) Response {
	r := Response{Description: desc}
	if len(example) > 0 {
		r.Example = example[0]
	}
	return r
}

// OK200 返回 200 成功响应
func OK200(example interface{}) map[string]Response {
	return map[string]Response{"200": R("成功", example)}
}

// Err400 返回 400 参数错误响应
func Err400() Response {
	return R("参数错误", map[string]interface{}{
		"code":    400,
		"message": "参数错误",
	})
}

// Err401 返回 401 未授权响应
func Err401() Response {
	return R("未授权", map[string]interface{}{
		"code":    401,
		"message": "请先登录",
	})
}

// Err404 返回 404 资源不存在响应
func Err404() Response {
	return R("资源不存在", map[string]interface{}{
		"code":    404,
		"message": "资源不存在",
	})
}

// ============================================================
// 生成 JSON
// ============================================================

// BuildJSON 生成 Swagger 文档的 JSON 字节
func BuildJSON() ([]byte, error) {
	docMu.RLock()
	defer docMu.RUnlock()

	return json.MarshalIndent(doc, "", "  ")
}

// GetDoc 获取当前文档的副本
func GetDoc() Swagger {
	docMu.RLock()
	defer docMu.RUnlock()

	data, _ := json.Marshal(doc)
	var copy Swagger
	json.Unmarshal(data, &copy)
	return copy
}

// ============================================================
// 路由：暴露 JSON 和 Swagger UI
// ============================================================

// swaggerUIHTML 是一个内嵌的轻量级 Swagger UI HTML（从 unpkg CDN 加载静态资源）
const swaggerUIHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <title>Swagger UI - Lychee-Go</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui.css">
  <style>
    html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin: 0; background: #fafafa; }
    .topbar { display: none; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-bundle.js" crossorigin></script>
  <script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-standalone-preset.js" crossorigin></script>
  <script>
    window.onload = function() {
      SwaggerUIBundle({
        url: "__JSON_URL__",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
        plugins: [SwaggerUIBundle.plugins.DownloadUrl],
        layout: "StandaloneLayout"
      });
    };
  </script>
</body>
</html>`

// RegisterRoutes 在 Gin 中注册 Swagger 的路由（/swagger + /swagger/doc.json）
// 用法：在 route.Register() 中调用 swagger.RegisterRoutes(r)
func RegisterRoutes(r *gin.Engine) {
	basePath := config.GetString("swagger.path", "/swagger")
	if basePath == "" {
		basePath = "/swagger"
	}

	// 1. JSON 文档接口
	r.GET(path.Join(basePath, "doc.json"), func(c *gin.Context) {
		data, err := BuildJSON()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Data(http.StatusOK, "application/json; charset=utf-8", data)
	})

	// 2. 便捷 JSON 别名
	r.GET(path.Join(basePath, "json"), func(c *gin.Context) {
		data, err := BuildJSON()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Data(http.StatusOK, "application/json; charset=utf-8", data)
	})

	// 3. Swagger UI HTML 页面
	jsonURL := path.Join(basePath, "doc.json")
	html := swaggerUIHTML
	// 替换占位符
	html = replaceAll(html, "__JSON_URL__", jsonURL)
	htmlBytes := []byte(html)

	r.GET(basePath, func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", htmlBytes)
	})
	r.GET(basePath+"/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", htmlBytes)
	})

	logger.Info("[Swagger] Routes registered: %s, %s",
		basePath, path.Join(basePath, "doc.json"))
}

// ============================================================
// 自动扫描 Gin 路由生成 Swagger（可选高级能力）
// ============================================================

// ScanGinRoutes 扫描已注册到 Gin 的路由，自动补充基础的 Swagger 定义
// 注意：扫描的路由只包含 method 和 path，不含参数和响应示例，需要手动补全
func ScanGinRoutes(r *gin.Engine) {
	routes := r.Routes()
	added := 0
	for _, rt := range routes {
		// 跳过自身路由
		if rt.Path == "/swagger/doc.json" || rt.Path == "/swagger" || rt.Path == "/swagger/json" {
			continue
		}
		docMu.Lock()
		if doc.Paths == nil {
			doc.Paths = make(map[string]map[string]Endpoint)
		}
		if _, exists := doc.Paths[rt.Path]; !exists {
			doc.Paths[rt.Path] = make(map[string]Endpoint)
		}
		m := stringsToLower(rt.Method)
		if _, exists := doc.Paths[rt.Path][m]; !exists {
			doc.Paths[rt.Path][m] = Endpoint{
				Summary:     fmt.Sprintf("[%s] %s", rt.Method, rt.Path),
				Description: "Auto-generated from registered Gin routes",
				Responses: map[string]Response{
					"200": R("成功"),
				},
			}
			added++
		}
		docMu.Unlock()
	}
	if added > 0 {
		logger.Info("[Swagger] Auto-scanned %d routes from Gin", added)
	}
}

// PrintSummary 在终端打印已注册的端点摘要
func PrintSummary() {
	docMu.RLock()
	defer docMu.RUnlock()

	paths := make([]string, 0, len(doc.Paths))
	for p := range doc.Paths {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	fmt.Println()
	fmt.Println("========================================================")
	fmt.Printf(" Swagger API Documentation (%d endpoints)\n", len(doc.Paths))
	fmt.Println("========================================================")
	for _, p := range paths {
		methods := []string{}
		for m := range doc.Paths[p] {
			methods = append(methods, m)
		}
		sort.Strings(methods)
		for _, m := range methods {
			fmt.Printf("  %-7s %s\n", stringsToUpper(m), p)
		}
	}
	fmt.Println("========================================================")
	fmt.Println(" Swagger UI:   http://localhost" + fmt.Sprintf(":%d", config.GetInt("app.port", 8080)) + config.GetString("swagger.path", "/swagger"))
	fmt.Println(" JSON Schema:  " + "http://localhost" + fmt.Sprintf(":%d", config.GetInt("app.port", 8080)) + config.GetString("swagger.path", "/swagger") + "/doc.json")
	fmt.Println("========================================================")
	fmt.Println()
}

// ============================================================
// 小工具
// ============================================================

func stringsToLower(s string) string {
	out := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		out[i] = c
	}
	return string(out)
}

func stringsToUpper(s string) string {
	out := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			c -= 32
		}
		out[i] = c
	}
	return string(out)
}

func replaceAll(s, old, new string) string {
	result := s
	for {
		pos := indexOf(result, old)
		if pos < 0 {
			break
		}
		result = result[:pos] + new + result[pos+len(old):]
	}
	return result
}

func indexOf(s, sub string) int {
	n := len(sub)
	if n == 0 {
		return 0
	}
	for i := 0; i+n <= len(s); i++ {
		if s[i:i+n] == sub {
			return i
		}
	}
	return -1
}
