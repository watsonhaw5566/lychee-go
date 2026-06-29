# filesystem · 文件系统

统一的文件操作接口，支持本地磁盘、阿里云 OSS、腾讯云 COS，可扩展其他云存储。

## 支持的驱动

| 驱动 | 说明 | 状态 |
|------|------|------|
| local | 本地文件系统 | ✅ 已实现 |
| oss | 阿里云 OSS | ✅ 已实现 |
| cos | 腾讯云 COS | ✅ 已实现 |

## 用法

```go
import "lychee-go/internal/filesystem"

// ---------- 使用默认驱动 ----------
filesystem.Put("users/123/avatar.png", fileBytes)    // 写入文件
data, err := filesystem.Get("users/123/avatar.png")  // 读取文件
filesystem.Delete("users/123/avatar.png")            // 删除文件
exists := filesystem.Exists("files/test.txt")         // 是否存在
size, err := filesystem.Size("hello.txt")             // 文件大小（字节）
url := filesystem.URL("users/123/avatar.png")        // 访问 URL

// ---------- 复制本地文件到存储 ----------
err := filesystem.PutFile("uploads/backup.zip", "/tmp/local-file.zip")

// ---------- 指定驱动操作 ----------
ossDriver, err := filesystem.Disk("oss")
ossDriver.Put("backup/data.json", []byte(`{"key": "value"}`))

cosDriver, err := filesystem.Disk("cos")
content, err := cosDriver.Get("archive/2024.zip")

// ---------- 辅助工具 ----------
ext := filesystem.GetFileExt("image.png")           // 获取扩展名：png
filename := filesystem.GenerateFilename("avatar.jpg") // 生成安全文件名
allowed := filesystem.IsAllowedExt("test.php", []string{"jpg", "png", "gif"}) // 校验扩展名
```

## 配置

```yaml
filesystem:
  default: local  # 默认驱动：local / oss / cos

  # 本地驱动
  local:
    root: runtime/uploads       # 本地根目录（相对项目根目录）
    url: /uploads               # 访问 URL 前缀

  # 阿里云 OSS 驱动
  oss:
    enabled: false              # 是否启用
    access_key_id: your-access-key-id
    access_key_secret: your-access-key-secret
    bucket: your-bucket-name
    endpoint: oss-cn-hangzhou.aliyuncs.com
    url: https://your-bucket-name.oss-cn-hangzhou.aliyuncs.com

  # 腾讯云 COS 驱动
  cos:
    enabled: false              # 是否启用
    secret_id: your-secret-id
    secret_key: your-secret-key
    region: ap-guangzhou
    bucket: your-bucket-name
    url: https://your-bucket-name.cos.ap-guangzhou.myqcloud.com
```

## Driver 接口

实现以下接口即可扩展新驱动：

```go
type Driver interface {
    Put(path string, content []byte) error    // 写入文件内容
    PutFile(path string, srcFile string) error // 从本地文件复制
    Get(path string) ([]byte, error)          // 读取文件内容
    Delete(path string) error                 // 删除文件
    Exists(path string) bool                  // 判断是否存在
    Size(path string) (int64, error)          // 获取文件大小（字节）
    URL(path string) string                   // 获取访问 URL
    Path(path string) string                  // 获取存储路径标识
}
```

## 扩展新驱动

1. 实现 `Driver` 接口
2. 在 `Init()` 函数中注册驱动：
   ```go
   fs.Register("custom", NewCustomDriver(...))
   ```
3. 在 `config.yml` 中配置驱动参数
4. 设置 `filesystem.default` 切换默认驱动

## 注意事项

1. 使用云存储驱动前，需要先在配置文件中启用并填写正确的密钥信息
2. 路径参数不需要以 `/` 开头，内部会自动处理
3. 云存储的 `Path()` 方法返回虚拟路径标识（如 `oss://bucket/path`），仅供参考
4. 建议在生产环境使用云存储驱动，并配置 CDN 加速