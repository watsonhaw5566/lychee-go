package view

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"lychee-go/pkg/framework/config"
	"lychee-go/pkg/framework/logger"
)

type Config struct {
	ViewPath     string           // 模板文件目录
	CacheEnabled bool             // 是否启用模板缓存（生产环境建议开启）
	Extension    string           // 模板文件扩展名，默认为 .html
	Delimiters   []string         // 模板分隔符，默认 {{ }}
	FuncMap      template.FuncMap // 自定义模板函数
	Layout       string           // 全局布局模板
	ContentKey   string           // 内容区块的 key，默认为 "content"
}

var (
	defaultConfig = &Config{
		ViewPath:     "resources/views",
		CacheEnabled: false,
		Extension:    ".html",
		Delimiters:   []string{"{{", "}}"},
		FuncMap:      make(template.FuncMap),
		Layout:       "",
		ContentKey:   "content",
	}

	cache      = make(map[string]*template.Template)
	cacheMutex sync.RWMutex
	once       sync.Once
)

func Init() {
	once.Do(func() {
		defaultConfig.ViewPath = config.GetString("view.path", "resources/views")
		defaultConfig.CacheEnabled = config.GetBool("view.cache", false)
		defaultConfig.Extension = config.GetString("view.extension", ".html")
		defaultConfig.Layout = config.GetString("view.layout", "")
		defaultConfig.ContentKey = config.GetString("view.content_key", "content")

		initBuiltinFunctions()

		logger.Info("[View] Initialized (path: %s, cache: %v)",
			defaultConfig.ViewPath, defaultConfig.CacheEnabled)
	})
}

func initBuiltinFunctions() {
	defaultConfig.FuncMap["upper"] = strings.ToUpper
	defaultConfig.FuncMap["lower"] = strings.ToLower
	defaultConfig.FuncMap["title"] = strings.Title
	defaultConfig.FuncMap["trim"] = strings.TrimSpace
	defaultConfig.FuncMap["trimLeft"] = func(s, cutset string) string {
		return strings.TrimLeft(s, cutset)
	}
	defaultConfig.FuncMap["trimRight"] = func(s, cutset string) string {
		return strings.TrimRight(s, cutset)
	}
	defaultConfig.FuncMap["replace"] = strings.Replace
	defaultConfig.FuncMap["split"] = strings.Split
	defaultConfig.FuncMap["join"] = strings.Join
	defaultConfig.FuncMap["contains"] = strings.Contains
	defaultConfig.FuncMap["hasPrefix"] = strings.HasPrefix
	defaultConfig.FuncMap["hasSuffix"] = strings.HasSuffix
	defaultConfig.FuncMap["len"] = func(s string) int { return len(s) }
	defaultConfig.FuncMap["substr"] = func(s string, start, length int) string {
		if start < 0 {
			start = 0
		}
		if start >= len(s) {
			return ""
		}
		if length <= 0 || start+length > len(s) {
			return s[start:]
		}
		return s[start : start+length]
	}

	defaultConfig.FuncMap["add"] = func(a, b int) int {
		return a + b
	}
	defaultConfig.FuncMap["sub"] = func(a, b int) int {
		return a - b
	}
	defaultConfig.FuncMap["mul"] = func(a, b int) int {
		return a * b
	}
	defaultConfig.FuncMap["div"] = func(a, b int) (int, error) {
		if b == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return a / b, nil
	}
	defaultConfig.FuncMap["mod"] = func(a, b int) (int, error) {
		if b == 0 {
			return 0, fmt.Errorf("mod by zero")
		}
		return a % b, nil
	}

	defaultConfig.FuncMap["eq"] = func(a, b interface{}) bool {
		return a == b
	}
	defaultConfig.FuncMap["ne"] = func(a, b interface{}) bool {
		return a != b
	}
	defaultConfig.FuncMap["gt"] = func(a, b int) bool {
		return a > b
	}
	defaultConfig.FuncMap["lt"] = func(a, b int) bool {
		return a < b
	}
	defaultConfig.FuncMap["ge"] = func(a, b int) bool {
		return a >= b
	}
	defaultConfig.FuncMap["le"] = func(a, b int) bool {
		return a <= b
	}
	defaultConfig.FuncMap["and"] = func(a, b bool) bool {
		return a && b
	}
	defaultConfig.FuncMap["or"] = func(a, b bool) bool {
		return a || b
	}
	defaultConfig.FuncMap["not"] = func(a bool) bool {
		return !a
	}

	defaultConfig.FuncMap["printf"] = fmt.Sprintf
	defaultConfig.FuncMap["html"] = func(s string) template.HTML {
		return template.HTML(s)
	}
	defaultConfig.FuncMap["urlquery"] = func(s string) template.URL {
		return template.URL(url.QueryEscape(s))
	}

	defaultConfig.FuncMap["include"] = func(name string, data interface{}) (template.HTML, error) {
		var buf bytes.Buffer
		err := Render(&buf, name, data)
		if err != nil {
			return "", err
		}
		return template.HTML(buf.String()), nil
	}
}

func AddFunc(name string, fn interface{}) {
	defaultConfig.FuncMap[name] = fn
	logger.Info("[View] Registered custom function: %s", name)
}

func AddFuncs(funcs template.FuncMap) {
	for name, fn := range funcs {
		defaultConfig.FuncMap[name] = fn
	}
	logger.Info("[View] Registered %d custom functions", len(funcs))
}

func Render(w io.Writer, templateName string, data interface{}) error {
	tpl, err := getTemplate(templateName)
	if err != nil {
		return fmt.Errorf("template load failed: %w", err)
	}

	return tpl.Execute(w, data)
}

func RenderString(templateName string, data interface{}) (string, error) {
	var buf bytes.Buffer
	err := Render(&buf, templateName, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func RenderWithLayout(w io.Writer, templateName, layoutName string, data interface{}) error {
	content, err := RenderString(templateName, data)
	if err != nil {
		return err
	}

	layoutData := make(map[string]interface{})
	if d, ok := data.(map[string]interface{}); ok {
		for k, v := range d {
			layoutData[k] = v
		}
	}
	layoutData[defaultConfig.ContentKey] = template.HTML(content)

	return Render(w, layoutName, layoutData)
}

func getTemplate(name string) (*template.Template, error) {
	if !strings.HasSuffix(name, defaultConfig.Extension) {
		name += defaultConfig.Extension
	}

	if defaultConfig.CacheEnabled {
		cacheMutex.RLock()
		if tpl, ok := cache[name]; ok {
			cacheMutex.RUnlock()
			return tpl, nil
		}
		cacheMutex.RUnlock()
	}

	templatePath := filepath.Join(defaultConfig.ViewPath, name)
	if !fileExists(templatePath) {
		return nil, fmt.Errorf("template file not found: %s", templatePath)
	}

	tpl := template.New(name).Delims(defaultConfig.Delimiters[0], defaultConfig.Delimiters[1])
	tpl = tpl.Funcs(defaultConfig.FuncMap)

	tpl, err := tpl.ParseFiles(templatePath)
	if err != nil {
		return nil, fmt.Errorf("template parse failed: %w", err)
	}

	if defaultConfig.Layout != "" {
		layoutPath := filepath.Join(defaultConfig.ViewPath, defaultConfig.Layout)
		if !strings.HasSuffix(layoutPath, defaultConfig.Extension) {
			layoutPath += defaultConfig.Extension
		}
		if fileExists(layoutPath) {
			tpl, err = tpl.ParseFiles(layoutPath)
			if err != nil {
				return nil, fmt.Errorf("layout parse failed: %w", err)
			}
		}
	}

	if defaultConfig.CacheEnabled {
		cacheMutex.Lock()
		cache[name] = tpl
		cacheMutex.Unlock()
	}

	return tpl, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func ClearCache() {
	cacheMutex.Lock()
	cache = make(map[string]*template.Template)
	cacheMutex.Unlock()
	logger.Info("[View] Template cache cleared")
}

func GetCacheSize() int {
	cacheMutex.RLock()
	size := len(cache)
	cacheMutex.RUnlock()
	return size
}

func GetConfig() *Config {
	return defaultConfig
}

func SetConfig(cfg *Config) {
	if cfg == nil {
		return
	}
	if cfg.ViewPath != "" {
		defaultConfig.ViewPath = cfg.ViewPath
	}
	defaultConfig.CacheEnabled = cfg.CacheEnabled
	if cfg.Extension != "" {
		defaultConfig.Extension = cfg.Extension
	}
	if len(cfg.Delimiters) == 2 {
		defaultConfig.Delimiters = cfg.Delimiters
	}
	if len(cfg.FuncMap) > 0 {
		for k, v := range cfg.FuncMap {
			defaultConfig.FuncMap[k] = v
		}
	}
	if cfg.Layout != "" {
		defaultConfig.Layout = cfg.Layout
	}
	if cfg.ContentKey != "" {
		defaultConfig.ContentKey = cfg.ContentKey
	}
}

func ListTemplates() ([]string, error) {
	var templates []string
	err := filepath.Walk(defaultConfig.ViewPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, defaultConfig.Extension) {
			relPath, err := filepath.Rel(defaultConfig.ViewPath, path)
			if err != nil {
				return err
			}
			templates = append(templates, strings.TrimSuffix(relPath, defaultConfig.Extension))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return templates, nil
}

func TemplateExists(name string) bool {
	if !strings.HasSuffix(name, defaultConfig.Extension) {
		name += defaultConfig.Extension
	}
	return fileExists(filepath.Join(defaultConfig.ViewPath, name))
}
