package utils

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type HTTPClient struct {
	client *http.Client
}

func NewHTTPClient(timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *HTTPClient) Get(url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.client.Do(req)
}

func ParseStatusCodes(codes string) []int {
	var result []int
	for _, code := range strings.Split(codes, ",") {
		if statusCode := ParseInt(strings.TrimSpace(code)); statusCode > 0 {
			result = append(result, statusCode)
		}
	}
	return result
}

func ParseInt(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}

func ParseDuration(s string) (time.Duration, error) {
	if _, err := fmt.Sscanf(s, "%d", new(int)); err == nil {
		s += "s"
	}
	return time.ParseDuration(s)
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func WriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func AbsPath(path string) string {
	wd, err := os.Getwd()
	if err != nil {
		return path // Fallback to relative path if Getwd fails
	}
	return wd + "/" + path
}

func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func HashMD5(data []byte) string {
	h := md5.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}