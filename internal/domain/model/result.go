package model

import (
	"time"
)

// ScanResult 扫描结果类型，与Result保持一致
type ScanResult struct {
	URL           string            `json:"url"`
	StatusCode    int               `json:"status_code"`
	ContentLength int64             `json:"content_length"`
	Headers       map[string]string `json:"headers"`
	Body          string            `json:"body,omitempty"`
	Error         string            `json:"error,omitempty"`
	Time          time.Time         `json:"time"`
	ResponseTime  time.Duration     `json:"response_time"`
	TechStack     []string          `json:"tech_stack,omitempty"`
	BodyHash      string            `json:"body_hash,omitempty"` // 响应体hash
}