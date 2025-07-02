package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/diraiscan/internal/pkg/config"
	"github.com/diraiscan/internal/pkg/utils"
	"github.com/diraiscan/internal/domain/model"

	"go.uber.org/zap"
)

// Analysis AI分析结果
type Analysis struct {
	URL         string    `json:"url"`
	Risk        string    `json:"risk"`        // 风险等级：low/medium/high
	Confidence  float64   `json:"confidence"`  // 置信度
	Description string    `json:"description"` // 分析描述
	Time        time.Time `json:"time"`
}

// Analyzer AI分析器
type Analyzer struct {
	config *config.Config
	client *http.Client
}

// NewAnalyzer 创建新的AI分析器
func NewAnalyzer(cfg *config.Config) *Analyzer {
	// 解析AI超时时间
	timeout, err := time.ParseDuration(cfg.AI.Timeout)
	if err != nil {
		timeout = 30 * time.Second // 默认30秒
	}

	return &Analyzer{
		config: cfg,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Analyze 分析扫描结果
func (a *Analyzer) Analyze(results []model.Result) ([]Analysis, error) {
	if !a.config.AI.Enabled {
		return nil, fmt.Errorf("AI功能未启用")
	}

	analyses := make([]Analysis, 0)
	for _, result := range results {
		// 跳过错误结果
		if result.Error != "" {
			continue
		}

		// 分析单个结果
		analysis, err := a.analyzeResult(result)
		if err != nil {
			utils.Error("AI分析失败",
				zap.String("url", result.URL),
				zap.Error(err),
			)
			continue
		}

		analyses = append(analyses, analysis)
	}

	return analyses, nil
}

// analyzeResult 分析单个扫描结果
func (a *Analyzer) analyzeResult(result model.Result) (Analysis, error) {
	// 准备请求数据
	reqData := map[string]interface{}{
		"url":            result.URL,
		"status_code":    result.StatusCode,
		"content_length": result.ContentLength,
		"headers":        result.Headers,
		"body":           result.Body,
	}

	data, err := json.Marshal(reqData)
	if err != nil {
		return Analysis{}, fmt.Errorf("序列化请求数据失败: %v", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", a.config.AI.Endpoint, bytes.NewBuffer(data))
	if err != nil {
		return Analysis{}, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	// 如果需要API Key，可以在这里添加
	// req.Header.Set("Authorization", "Bearer "+apiKey)

	// 发送请求
	resp, err := a.client.Do(req)
	if err != nil {
		return Analysis{}, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var analysis Analysis
	if err := json.NewDecoder(resp.Body).Decode(&analysis); err != nil {
		return Analysis{}, fmt.Errorf("解析响应失败: %v", err)
	}

	analysis.Time = time.Now()
	return analysis, nil
}
