package filesystem

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"lychee-go/pkg/framework/config"
	"lychee-go/pkg/framework/logger"
)

// ======== 文件驱动接口 ========

type Driver interface {
	// Put 写入文件
	Put(path string, content []byte) error
	// PutFile 从本地文件复制
	PutFile(path string, srcFile string) error
	// Get 读取文件
	Get(path string) ([]byte, error)
	// Delete 删除文件
	Delete(path string) error
	// Exists 判断是否存在
	Exists(path string) bool
	// Size 获取文件大小（字节）
	Size(path string) (int64, error)
	// URL 获取访问 URL
	URL(path string) string
	// Path 获取本地绝对路径（本地驱动有效）
	Path(path string) string
}

// ======== 本地驱动 ========

type LocalDriver struct {
	rootDir string // 根目录
	baseURL string // 访问 URL 前缀
}

func NewLocalDriver(rootDir, baseURL string) *LocalDriver {
	return &LocalDriver{
		rootDir: strings.TrimRight(rootDir, "/"),
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

func (d *LocalDriver) fullPath(path string) string {
	// 防止路径穿越攻击
	path = strings.TrimLeft(path, "/")
	path = strings.ReplaceAll(path, "..", "")
	return d.rootDir + "/" + path
}

func (d *LocalDriver) Put(path string, content []byte) error {
	fullPath := d.fullPath(path)

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("mkdir failed: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(fullPath, content, 0644); err != nil {
		return fmt.Errorf("write file failed: %w", err)
	}

	logger.Debug("[Filesystem] Put file: %s (%d bytes)", fullPath, len(content))
	return nil
}

func (d *LocalDriver) PutFile(path string, srcFile string) error {
	src, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer src.Close()

	fullPath := d.fullPath(path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("mkdir failed: %w", err)
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	logger.Debug("[Filesystem] PutFile: %s -> %s", srcFile, fullPath)
	return err
}

func (d *LocalDriver) Get(path string) ([]byte, error) {
	return os.ReadFile(d.fullPath(path))
}

func (d *LocalDriver) Delete(path string) error {
	fullPath := d.fullPath(path)
	if !d.Exists(path) {
		return nil
	}
	err := os.Remove(fullPath)
	logger.Debug("[Filesystem] Delete: %s", fullPath)
	return err
}

func (d *LocalDriver) Exists(path string) bool {
	_, err := os.Stat(d.fullPath(path))
	return err == nil
}

func (d *LocalDriver) Size(path string) (int64, error) {
	info, err := os.Stat(d.fullPath(path))
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func (d *LocalDriver) URL(path string) string {
	return d.baseURL + "/" + strings.TrimLeft(path, "/")
}

func (d *LocalDriver) Path(path string) string {
	return d.fullPath(path)
}

// ======== 管理器 ========

type Manager struct {
	defaultDriver string
	drivers       map[string]Driver
}

func NewManager() *Manager {
	return &Manager{
		drivers: make(map[string]Driver),
	}
}

func (m *Manager) Register(name string, driver Driver) {
	m.drivers[name] = driver
	if m.defaultDriver == "" {
		m.defaultDriver = name
	}
	logger.Info("[Filesystem] Driver registered: %s", name)
}

func (m *Manager) SetDefault(name string) {
	m.defaultDriver = name
}

func (m *Manager) Driver(name ...string) (Driver, error) {
	driverName := m.defaultDriver
	if len(name) > 0 {
		driverName = name[0]
	}

	driver, ok := m.drivers[driverName]
	if !ok {
		return nil, fmt.Errorf("filesystem driver not found: %s", driverName)
	}
	return driver, nil
}

// ======== 全局实例 ========

var fs *Manager

// Init 初始化文件系统（从 config.yml 读取配置）
func Init() error {
	fs = NewManager()

	defaultDriver := config.GetString("filesystem.default", "local")

	// 本地驱动
	localRoot := config.GetString("filesystem.local.root", "runtime/uploads")
	localURL := config.GetString("filesystem.local.url", "/uploads")
	fs.Register("local", NewLocalDriver(localRoot, localURL))

	// 阿里云 OSS 驱动
	ossEnabled := config.GetBool("filesystem.oss.enabled", false)
	if ossEnabled {
		ossEndpoint := config.GetString("filesystem.oss.endpoint", "")
		ossAccessKeyID := config.GetString("filesystem.oss.access_key_id", "")
		ossAccessKeySecret := config.GetString("filesystem.oss.access_key_secret", "")
		ossBucket := config.GetString("filesystem.oss.bucket", "")
		ossBaseURL := config.GetString("filesystem.oss.url", "")
		if ossEndpoint != "" && ossAccessKeyID != "" && ossAccessKeySecret != "" && ossBucket != "" {
			ossDriver, err := NewOSSDriver(ossEndpoint, ossAccessKeyID, ossAccessKeySecret, ossBucket, ossBaseURL)
			if err != nil {
				logger.Warn("[Filesystem] OSS driver init failed: %v", err)
			} else {
				fs.Register("oss", ossDriver)
			}
		}
	}

	// 腾讯云 COS 驱动
	cosEnabled := config.GetBool("filesystem.cos.enabled", false)
	if cosEnabled {
		cosSecretID := config.GetString("filesystem.cos.secret_id", "")
		cosSecretKey := config.GetString("filesystem.cos.secret_key", "")
		cosRegion := config.GetString("filesystem.cos.region", "")
		cosBucket := config.GetString("filesystem.cos.bucket", "")
		cosBaseURL := config.GetString("filesystem.cos.url", "")
		if cosSecretID != "" && cosSecretKey != "" && cosRegion != "" && cosBucket != "" {
			cosDriver, err := NewCOSDriver(cosSecretID, cosSecretKey, cosRegion, cosBucket, cosBaseURL)
			if err != nil {
				logger.Warn("[Filesystem] COS driver init failed: %v", err)
			} else {
				fs.Register("cos", cosDriver)
			}
		}
	}

	fs.SetDefault(defaultDriver)
	logger.Info("[Filesystem] Initialized (default: %s)", defaultDriver)
	return nil
}

// ======== 对外 API（类似 ThinkPHP 的 Storage::disk()） ========

// Disk 获取指定驱动
func Disk(name ...string) (Driver, error) {
	if fs == nil {
		return nil, errors.New("filesystem not initialized, call filesystem.Init() first")
	}
	return fs.Driver(name...)
}

// Put 写入文件（默认驱动）
func Put(path string, content []byte) error {
	d, err := Disk()
	if err != nil {
		return err
	}
	return d.Put(path, content)
}

// PutFile 从源文件复制
func PutFile(path string, srcFile string) error {
	d, err := Disk()
	if err != nil {
		return err
	}
	return d.PutFile(path, srcFile)
}

// Get 读取文件
func Get(path string) ([]byte, error) {
	d, err := Disk()
	if err != nil {
		return nil, err
	}
	return d.Get(path)
}

// Delete 删除
func Delete(path string) error {
	d, err := Disk()
	if err != nil {
		return err
	}
	return d.Delete(path)
}

// Exists 判断存在
func Exists(path string) bool {
	d, err := Disk()
	if err != nil {
		return false
	}
	return d.Exists(path)
}

// URL 获取访问链接
func URL(path string) string {
	d, err := Disk()
	if err != nil {
		return path
	}
	return d.URL(path)
}

// ======== 辅助工具 ========

// GetFileExt 获取文件扩展名（小写，不带点）
func GetFileExt(filename string) string {
	ext := filepath.Ext(filename)
	return strings.ToLower(strings.TrimPrefix(ext, "."))
}

// GenerateFilename 生成安全文件名（原扩展名 + 时间戳）
func GenerateFilename(originalName string) string {
	ext := GetFileExt(originalName)
	if ext == "" {
		ext = "bin"
	}
	timestamp := time.Now().Format("20060102150405")
	return fmt.Sprintf("%s_%d.%s", timestamp, time.Now().UnixNano()%1000000, ext)
}

// IsAllowedExt 校验文件扩展名是否在白名单中
func IsAllowedExt(filename string, allowedExts []string) bool {
	ext := GetFileExt(filename)
	for _, allowed := range allowedExts {
		if strings.ToLower(allowed) == ext {
			return true
		}
	}
	return false
}
