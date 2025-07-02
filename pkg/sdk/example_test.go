package sdk_test

import (
	"fmt"
	"time"

	"github.com/diraiscan/pkg/sdk"
)

func Example() {
	scanner := sdk.NewScanner()

	scanner.SetOptions(&sdk.ScanOptions{
		Threads:    50,
		Timeout:    time.Second * 30,
		RetryTimes: 3,
		RetryDelay: time.Second * 5,
	})

	targets := []string{"example.com", "test.com"}
	results, err := scanner.Scan(targets)
	if err != nil {
		fmt.Printf("扫描错误: %v\n", err)
		return
	}

	analyzer := sdk.NewAnalyzer()

	analyzer.SetOptions(&sdk.AIOptions{
		Enabled:  true,
		Model:    "qwen3:4b",
		Endpoint: "http://localhost:11434",
		Local:    true,
	})

	analyses, err := analyzer.Analyze(results)
	if err != nil {
		fmt.Printf("分析错误: %v\n", err)
		return
	}

	for _, analysis := range analyses {
		fmt.Printf("URL: %s\n", analysis.OriginalResult.URL)
		fmt.Printf("风险等级: %s\n", analysis.RiskLevel)
		fmt.Printf("发现: %v\n", analysis.Findings)
		fmt.Printf("建议: %v\n", analysis.Suggestions)
	}
}
