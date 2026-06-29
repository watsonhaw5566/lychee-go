package filesystem

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"lychee-go/pkg/framework/logger"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type OSSDriver struct {
	client     *oss.Client
	bucket     *oss.Bucket
	bucketName string
	endpoint   string
	baseURL    string
}

func NewOSSDriver(endpoint, accessKeyID, accessKeySecret, bucketName, baseURL string) (*OSSDriver, error) {
	client, err := oss.New(endpoint, accessKeyID, accessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("oss client create failed: %w", err)
	}

	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return nil, fmt.Errorf("oss bucket get failed: %w", err)
	}

	return &OSSDriver{
		client:     client,
		bucket:     bucket,
		bucketName: bucketName,
		endpoint:   endpoint,
		baseURL:    strings.TrimRight(baseURL, "/"),
	}, nil
}

func (d *OSSDriver) Put(path string, content []byte) error {
	path = strings.TrimLeft(path, "/")
	err := d.bucket.PutObject(path, bytes.NewReader(content))
	if err != nil {
		return fmt.Errorf("oss put object failed: %w", err)
	}
	logger.Debug("[Filesystem] OSS Put: %s (%d bytes)", path, len(content))
	return nil
}

func (d *OSSDriver) PutFile(path string, srcFile string) error {
	path = strings.TrimLeft(path, "/")
	src, err := os.Open(srcFile)
	if err != nil {
		return fmt.Errorf("open source file failed: %w", err)
	}
	defer src.Close()

	err = d.bucket.PutObject(path, src)
	if err != nil {
		return fmt.Errorf("oss put file failed: %w", err)
	}
	logger.Debug("[Filesystem] OSS PutFile: %s -> %s", srcFile, path)
	return nil
}

func (d *OSSDriver) Get(path string) ([]byte, error) {
	path = strings.TrimLeft(path, "/")
	body, err := d.bucket.GetObject(path)
	if err != nil {
		return nil, fmt.Errorf("oss get object failed: %w", err)
	}
	defer body.Close()

	return io.ReadAll(body)
}

func (d *OSSDriver) Delete(path string) error {
	path = strings.TrimLeft(path, "/")
	if !d.Exists(path) {
		return nil
	}
	err := d.bucket.DeleteObject(path)
	if err != nil {
		return fmt.Errorf("oss delete object failed: %w", err)
	}
	logger.Debug("[Filesystem] OSS Delete: %s", path)
	return nil
}

func (d *OSSDriver) Exists(path string) bool {
	path = strings.TrimLeft(path, "/")
	exists, err := d.bucket.IsObjectExist(path)
	if err != nil {
		logger.Error("[Filesystem] OSS Exists check failed: %v", err)
		return false
	}
	return exists
}

func (d *OSSDriver) Size(path string) (int64, error) {
	path = strings.TrimLeft(path, "/")
	props, err := d.bucket.GetObjectDetailedMeta(path)
	if err != nil {
		return 0, fmt.Errorf("oss get object meta failed: %w", err)
	}
	contentLen := props.Get("Content-Length")
	if contentLen == "" {
		return 0, fmt.Errorf("oss get object meta: Content-Length not found")
	}
	return strconv.ParseInt(contentLen, 10, 64)
}

func (d *OSSDriver) URL(path string) string {
	return d.baseURL + "/" + strings.TrimLeft(path, "/")
}

func (d *OSSDriver) Path(path string) string {
	return "oss://" + d.bucketName + "/" + strings.TrimLeft(path, "/")
}
