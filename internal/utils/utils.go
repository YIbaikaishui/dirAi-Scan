package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// HTTPClient HTTP客户端工具
type HTTPClient struct {
	client *http.Client
}

// NewHTTPClient 创建HTTP客户端
func NewHTTPClient(timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Get 发送GET请求
func (c *HTTPClient) Get(url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 添加请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.client.Do(req)
}

// ParseStatusCodes 解析状态码列表
func ParseStatusCodes(codes string) []int {
	var result []int
	for _, code := range strings.Split(codes, ",") {
		if statusCode := ParseInt(strings.TrimSpace(code)); statusCode > 0 {
			result = append(result, statusCode)
		}
	}
	return result
}

// ParseInt 解析整数
func ParseInt(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}

// ParseDuration 解析时间间隔
func ParseDuration(s string) (time.Duration, error) {
	// 如果没有单位，默认为秒
	if _, err := fmt.Sscanf(s, "%d", new(int)); err == nil {
		s += "s"
	}
	return time.ParseDuration(s)
}

// EnsureDir 确保目录存在
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// WriteFile 写入文件
func WriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

// FileExists 检查文件是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// HashMD5 计算数据的md5哈希
func HashMD5(data []byte) string {
	h := md5.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
