# filesystem · 文件系统

统一的文件操作接口，支持本地磁盘，可扩展 OSS / S3 等云存储。

## 用法

```go
import "lychee-go/internal/filesystem"

// ---------- 基本操作 ----------
filesystem.Put("users/123/avatar.png", fileBytes)    // 写入
data, err := filesystem.Get("users/123/avatar.png")  // 读取
filesystem.Delete("users/123/avatar.png")            // 删除
filesystem.Exists("files/test.txt")                  // 是否存在

// ---------- 访问 URL ----------
url := filesystem.URL("users/123/avatar.png")
// 返回：/uploads/users/123/avatar.png（本地） 或 https://oss.xxx/...（云存储）

// ---------- 内容类型辅助 ----------
filesystem.PutString("hello.txt", "Hello, world!")     // 写入字符串
content, err := filesystem.GetString("hello.txt")      // 读取为字符串

// ---------- 文件大小 ----------
size, err := filesystem.Size("hello.txt")
```

## 配置

```yaml
filesystem:
  default: local
  local:
    root: runtime/uploads       # 本地根目录（相对项目根）
    url: /uploads               # 访问 URL 前缀（需要 Nginx 或静态目录映射）
  # oss:
  #   access_key_id: your-access-key
  #   access_key_secret: your-secret
  #   bucket: your-bucket
  #   endpoint: oss-cn-hangzhou.aliyuncs.com
```

## 扩展新驱动

实现 `Driver` 接口即可：

```go
type Driver interface {
    Put(path string, content []byte) error
    Get(path string) ([]byte, error)
    Delete(path string) error
    Exists(path string) bool
    URL(path string) string
    Size(path string) (int64, error)
}
```

在 `Init()` 中注册，然后在 `config.yml` 的 `filesystem.default` 中配置名称即可切换。
