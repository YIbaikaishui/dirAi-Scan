package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/diraiscan/pkg/sdk"
)

var (
	// 扫描选项
	threads    = flag.Int("threads", 100, "并发线程数")
	timeout    = flag.Int("timeout", 30, "超时时间(秒)")
	userAgent  = flag.String("user-agent", "", "自定义User-Agent")
	proxy      = flag.String("proxy", "", "代理设置 (例如: http://127.0.0.1:8080)")
	retryTimes = flag.Int("retry", 3, "重试次数")
	retryDelay = flag.Int("retry-delay", 5, "重试延迟(秒)")

	// AI选项
	enableAI   = flag.Bool("ai", false, "启用AI分析")
	aiModel    = flag.String("ai-model", "qwen3:4b", "AI模型名称")
	aiKey      = flag.String("ai-key", "", "AI API密钥")
	aiEndpoint = flag.String("ai-endpoint", "http://localhost:11434", "AI API端点")
	aiLocal    = flag.Bool("ai-local", true, "使用本地AI模型")

	// 输出选项
	outputFormat = flag.String("format", "text", "输出格式 (text/json)")
	outputFile   = flag.String("output", "", "输出文件路径")
)

func main() {
	// 解析命令行参数
	flag.Parse()

	// 获取目标列表
	targets := flag.Args()
	if len(targets) == 0 {
		fmt.Println("错误: 请提供扫描目标")
		flag.Usage()
		os.Exit(1)
	}

	// 创建扫描器
	scanner := sdk.NewScanner()

	// 配置扫描选项
	scanner.SetOptions(&sdk.ScanOptions{
		Threads:    *threads,
		Timeout:    time.Duration(*timeout) * time.Second,
		UserAgent:  *userAgent,
		Proxy:      *proxy,
		RetryTimes: *retryTimes,
		RetryDelay: time.Duration(*retryDelay) * time.Second,
	})

	// 执行扫描
	results, err := scanner.Scan(targets)
	if err != nil {
		fmt.Printf("扫描错误: %v\n", err)
		os.Exit(1)
	}

	// 如果启用AI分析
	if *enableAI {
		analyzer := sdk.NewAnalyzer()
		analyzer.SetOptions(&sdk.AIOptions{
			Enabled:  true,
			Model:    *aiModel,
			APIKey:   *aiKey,
			Endpoint: *aiEndpoint,
			Local:    *aiLocal,
		})

		// 分析结果
		analyses, err := analyzer.Analyze(results)
		if err != nil {
			fmt.Printf("AI分析错误: %v\n", err)
		} else {
			// 输出分析结果
			outputAnalyses(analyses)
		}
	} else {
		// 输出扫描结果
		outputResults(results)
	}
}

// outputResults 输出扫描结果
func outputResults(results []sdk.ScanResult) {
	for _, result := range results {
		fmt.Printf("URL: %s\n", result.URL)
		fmt.Printf("状态码: %d\n", result.StatusCode)
		fmt.Printf("内容长度: %d\n", result.ContentLength)
		fmt.Printf("响应时间: %s\n", result.ResponseTime)
		fmt.Printf("技术栈: %s\n\n", strings.Join(result.TechStack, ", "))
	}
}

// outputAnalyses 输出分析结果
func outputAnalyses(analyses []sdk.ScanResultAnalysis) {
	for _, analysis := range analyses {
		fmt.Printf("URL: %s\n", analysis.OriginalResult.URL)
		fmt.Printf("风险等级: %s\n", analysis.RiskLevel)
		fmt.Printf("发现:\n")
		for _, finding := range analysis.Findings {
			fmt.Printf("  - %s\n", finding)
		}
		fmt.Printf("建议:\n")
		for _, suggestion := range analysis.Suggestions {
			fmt.Printf("  - %s\n", suggestion)
		}
		fmt.Println()
	}
}
