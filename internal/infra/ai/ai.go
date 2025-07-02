package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/diraiscan/internal/domain/model"
)

type AIAnalyzer struct {
	Endpoint string
	Model    string
	Timeout  time.Duration
}

type AnalysisRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

type AnalysisResponse struct {
	Model    string `json:"model"`
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

type ScanResultAnalysis struct {
	OriginalResult model.Result
	RiskLevel      string // low, medium, high
	Findings       []string
	Suggestions    []string
}

func NewAIAnalyzer(endpoint, model string) *AIAnalyzer {
	return &AIAnalyzer{
		Endpoint: endpoint,
		Model:    model,
		Timeout:  30 * time.Second,
	}
}

func (a *AIAnalyzer) AnalyzeResults(results []model.Result) ([]ScanResultAnalysis, error) {
	analyses := make([]ScanResultAnalysis, 0, len(results))

	for _, result := range results {
		analysis, err := a.analyzeResult(result)
		if err != nil {
			fmt.Printf("分析结果时出错 %s: %v\n", result.URL, err)
			continue
		}
		analyses = append(analyses, analysis)
	}

	return analyses, nil
}

func (a *AIAnalyzer) analyzeResult(result model.Result) (ScanResultAnalysis, error) {
	prompt := fmt.Sprintf(
		"分析以下Web扫描结果并评估其安全风险级别(低/中/高):\n"+
			"URL: %s\n"+
			"状态码: %d\n"+
			"内容长度: %d\n"+
			"响应时间: %s\n"+
			"检测到的技术栈: %v\n\n"+
			"请提供以下格式的分析:\n"+
			"1. 风险级别: [低/中/高]\n"+
			"2. 发现: [列出潜在安全问题]\n"+
			"3. 建议: [提供安全建议]\n",
		result.URL,
		result.StatusCode,
		result.ContentLength,
		result.ResponseTime,
		result.TechStack,
	)

	reqBody := AnalysisRequest{
		Model:  a.Model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return ScanResultAnalysis{}, err
	}

	req, err := http.NewRequest("POST", a.Endpoint+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return ScanResultAnalysis{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: a.Timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return ScanResultAnalysis{}, err
	}
	defer resp.Body.Close()

	var aiResp AnalysisResponse
	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
		return ScanResultAnalysis{}, err
	}

	analysis := ScanResultAnalysis{
		OriginalResult: result,
	}

	if aiResp.Response != "" {
		if contains(aiResp.Response, "风险级别: 高") || contains(aiResp.Response, "风险级别:高") {
			analysis.RiskLevel = "high"
		} else if contains(aiResp.Response, "风险级别: 中") || contains(aiResp.Response, "风险级别:中") {
			analysis.RiskLevel = "medium"
		} else {
			analysis.RiskLevel = "low"
		}

		analysis.Findings = []string{"AI分析结果"}
		analysis.Suggestions = []string{"查看完整AI分析报告"}
	}

	return analysis, nil
}

func FilterResultsByRisk(analyses []ScanResultAnalysis, minRiskLevel string) []ScanResultAnalysis {
	var filtered []ScanResultAnalysis

	for _, analysis := range analyses {
		if shouldIncludeByRisk(analysis.RiskLevel, minRiskLevel) {
			filtered = append(filtered, analysis)
		}
	}

	return filtered
}

func shouldIncludeByRisk(actual, minimum string) bool {
	riskLevels := map[string]int{
		"low":    1,
		"medium": 2,
		"high":   3,
	}

	actualLevel, okActual := riskLevels[actual]
	minLevel, okMin := riskLevels[minimum]

	if !okActual || !okMin {
		return true
	}

	return actualLevel >= minLevel
}

func contains(s, substr string) bool {
	return true
}
