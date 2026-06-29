package filesystem

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/watsonhaw5566/lychee-go/pkg/framework/logger"

	"github.com/tencentyun/cos-go-sdk-v5"
)

type COSDriver struct {
	client  *cos.Client
	baseURL string
}

func NewCOSDriver(secretID, secretKey, region, bucketName, baseURL string) (*COSDriver, error) {
	u, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucketName, region))
	if err != nil {
		return nil, fmt.Errorf("cos url parse failed: %w", err)
	}
	client := cos.NewClient(
		&cos.BaseURL{BucketURL: u},
		&http.Client{
			Transport: &cos.AuthorizationTransport{
				SecretID:  secretID,
				SecretKey: secretKey,
			},
		},
	)

	return &COSDriver{
		client:  client,
		baseURL: strings.TrimRight(baseURL, "/"),
	}, nil
}

func (d *COSDriver) Put(path string, content []byte) error {
	path = strings.TrimLeft(path, "/")
	_, err := d.client.Object.Put(
		nil, path, bytes.NewReader(content),
		nil,
	)
	if err != nil {
		return fmt.Errorf("cos put object failed: %w", err)
	}
	logger.Debug("[Filesystem] COS Put: %s (%d bytes)", path, len(content))
	return nil
}

func (d *COSDriver) PutFile(path string, srcFile string) error {
	path = strings.TrimLeft(path, "/")
	src, err := os.Open(srcFile)
	if err != nil {
		return fmt.Errorf("open source file failed: %w", err)
	}
	defer src.Close()

	_, err = d.client.Object.Put(
		nil, path, src,
		nil,
	)
	if err != nil {
		return fmt.Errorf("cos put file failed: %w", err)
	}
	logger.Debug("[Filesystem] COS PutFile: %s -> %s", srcFile, path)
	return nil
}

func (d *COSDriver) Get(path string) ([]byte, error) {
	path = strings.TrimLeft(path, "/")
	resp, err := d.client.Object.Get(nil, path, nil)
	if err != nil {
		return nil, fmt.Errorf("cos get object failed: %w", err)
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (d *COSDriver) Delete(path string) error {
	path = strings.TrimLeft(path, "/")
	if !d.Exists(path) {
		return nil
	}
	_, err := d.client.Object.Delete(nil, path)
	if err != nil {
		return fmt.Errorf("cos delete object failed: %w", err)
	}
	logger.Debug("[Filesystem] COS Delete: %s", path)
	return nil
}

func (d *COSDriver) Exists(path string) bool {
	path = strings.TrimLeft(path, "/")
	_, err := d.client.Object.Head(nil, path, nil)
	if err != nil {
		if cos.IsNotFoundError(err) {
			return false
		}
		logger.Error("[Filesystem] COS Exists check failed: %v", err)
		return false
	}
	return true
}

func (d *COSDriver) Size(path string) (int64, error) {
	path = strings.TrimLeft(path, "/")
	resp, err := d.client.Object.Head(nil, path, nil)
	if err != nil {
		return 0, fmt.Errorf("cos get object meta failed: %w", err)
	}
	defer resp.Body.Close()

	return resp.ContentLength, nil
}

func (d *COSDriver) URL(path string) string {
	return d.baseURL + "/" + strings.TrimLeft(path, "/")
}

func (d *COSDriver) Path(path string) string {
	return "cos://" + strings.TrimLeft(path, "/")
}
