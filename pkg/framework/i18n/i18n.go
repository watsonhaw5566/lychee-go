package i18n

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/watsonhaw5566/lychee-go/pkg/framework/config"
	"github.com/watsonhaw5566/lychee-go/pkg/framework/logger"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

var (
	translations = make(map[string]map[string]string)
	defaultLang  = "zh-CN"
	currentLang  = "zh-CN"
	mu           sync.RWMutex
)

type I18n struct {
	lang string
}

func Init() {
	defaultLang = config.GetString("i18n.default", "zh-CN")
	currentLang = defaultLang

	dir := config.GetString("i18n.dir", "resources/lang")

	if err := loadTranslations(dir); err != nil {
		logger.Warn("[I18n] Load translations failed: %v", err)
	} else {
		logger.Info("[I18n] Initialized with default lang: %s, loaded %d languages", defaultLang, len(translations))
	}
}

func loadTranslations(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		logger.Warn("[I18n] Language directory not found: %s", dir)
		return nil
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.yml"))
	if err != nil {
		return fmt.Errorf("glob failed: %w", err)
	}

	for _, file := range files {
		lang := strings.TrimSuffix(filepath.Base(file), ".yml")
		if err := loadLanguageFile(lang, file); err != nil {
			logger.Warn("[I18n] Load %s failed: %v", file, err)
		}
	}

	return nil
}

func loadLanguageFile(lang, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file failed: %w", err)
	}

	var data map[string]string
	if err := yaml.Unmarshal(content, &data); err != nil {
		return fmt.Errorf("unmarshal failed: %w", err)
	}

	mu.Lock()
	translations[lang] = data
	mu.Unlock()

	logger.Debug("[I18n] Loaded %d translations for %s", len(data), lang)
	return nil
}

func SetDefaultLang(lang string) {
	mu.Lock()
	defaultLang = lang
	mu.Unlock()
}

func GetDefaultLang() string {
	mu.RLock()
	defer mu.RUnlock()
	return defaultLang
}

func SetLang(lang string) {
	mu.Lock()
	currentLang = lang
	mu.Unlock()
}

func GetLang() string {
	mu.RLock()
	defer mu.RUnlock()
	return currentLang
}

func Get(key string, args ...interface{}) string {
	return getWithLang(currentLang, key, args...)
}

func getWithLang(lang, key string, args ...interface{}) string {
	mu.RLock()
	defer mu.RUnlock()

	trans, ok := translations[lang]
	if !ok {
		trans, ok = translations[defaultLang]
		if !ok {
			return key
		}
	}

	value, ok := trans[key]
	if !ok {
		trans, ok = translations[defaultLang]
		if !ok {
			return key
		}
		value, ok = trans[key]
		if !ok {
			return key
		}
	}

	if len(args) > 0 {
		return fmt.Sprintf(value, args...)
	}
	return value
}

func T(key string, args ...interface{}) string {
	return Get(key, args...)
}

func ForLang(lang string) *I18n {
	return &I18n{lang: lang}
}

func (i *I18n) Get(key string, args ...interface{}) string {
	return getWithLang(i.lang, key, args...)
}

func (i *I18n) T(key string, args ...interface{}) string {
	return i.Get(key, args...)
}

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := detectLang(c)
		c.Set("lang", lang)
		c.Next()
	}
}

func detectLang(c *gin.Context) string {
	lang := c.GetHeader("Accept-Language")
	if lang != "" {
		lang = parseAcceptLanguage(lang)
		if isValidLang(lang) {
			return lang
		}
	}

	lang = c.Query("lang")
	if lang != "" && isValidLang(lang) {
		return lang
	}

	lang, _ = c.Cookie("lang")
	if lang != "" && isValidLang(lang) {
		return lang
	}

	return defaultLang
}

func parseAcceptLanguage(header string) string {
	parts := strings.Split(header, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		idx := strings.Index(part, ";")
		if idx != -1 {
			part = part[:idx]
		}
		if isValidLang(part) {
			return part
		}
		baseLang := strings.Split(part, "-")[0]
		if isValidLang(baseLang) {
			return baseLang
		}
	}
	return ""
}

func isValidLang(lang string) bool {
	mu.RLock()
	defer mu.RUnlock()
	_, ok := translations[lang]
	return ok
}

func GetAvailableLangs() []string {
	mu.RLock()
	defer mu.RUnlock()
	langs := make([]string, 0, len(translations))
	for lang := range translations {
		langs = append(langs, lang)
	}
	return langs
}

func AddTranslation(lang, key, value string) {
	mu.Lock()
	defer mu.Unlock()
	if translations[lang] == nil {
		translations[lang] = make(map[string]string)
	}
	translations[lang][key] = value
}

func GetFromContext(c *gin.Context, key string, args ...interface{}) string {
	lang, _ := c.Get("lang")
	if langStr, ok := lang.(string); ok && langStr != "" {
		return getWithLang(langStr, key, args...)
	}
	return Get(key, args...)
}

func Printf(buf *bytes.Buffer, key string, args ...interface{}) {
	buf.WriteString(Get(key, args...))
}

func Sprintf(key string, args ...interface{}) string {
	return fmt.Sprintf(Get(key, args...), args...)
}
