package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/diraiscan/pkg/ai"
	"github.com/diraiscan/pkg/scanner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/diraiscan/internal/config"
	"github.com/diraiscan/internal/utils"
)

// 扫描命令
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "开始扫描目标",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 获取配置
		cfg := config.GetConfig()

		// 获取命令行参数
		targetURL, _ := cmd.Flags().GetString("url")
		if targetURL == "" {
			return fmt.Errorf("目标URL不能为空")
		}

		// 确保URL格式正确
		if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
			targetURL = "http://" + targetURL
		}

		// 获取输出路径
		outputPath, _ := cmd.Flags().GetString("output")
		if outputPath == "" {
			outputPath = "results"
		}

		// 创建输出目录
		if err := os.MkdirAll(outputPath, 0755); err != nil {
			return fmt.Errorf("创建输出目录失败: %v", err)
		}

		// 创建扫描引擎
		engine := scanner.NewEngine(cfg)

		// 从文件读取扫描路径
		paths, err := readPathsFromFile("data/common.txt")
		if err != nil {
			return fmt.Errorf("读取扫描路径失败: %v", err)
		}

		// 开始扫描
		if err := engine.Start(targetURL, paths); err != nil {
			return fmt.Errorf("启动扫描失败: %v", err)
		}

		// 处理结果
		results := make([]scanner.Result, 0)

		// 创建一个goroutine来处理错误
		go func() {
			for err := range engine.GetErrors() {
				utils.Error("扫描错误", zap.Error(err))
			}
		}()

		// 实时处理扫描结果
		for result := range engine.GetResults() {
			results = append(results, result)

			// 根据状态码设置颜色
			statusColor := ""
			statusSymbol := "[-]"
			switch {
			case result.StatusCode >= 200 && result.StatusCode < 300:
				statusColor = "\033[32m" // 绿色
				statusSymbol = "[+]"
			case result.StatusCode >= 300 && result.StatusCode < 400:
				statusColor = "\033[36m" // 青色
				statusSymbol = "[*]"
			case result.StatusCode >= 400 && result.StatusCode < 500:
				statusColor = "\033[33m" // 黄色
				statusSymbol = "[-]"
			case result.StatusCode >= 500:
				statusColor = "\033[31m" // 红色
				statusSymbol = "[!]"
			}

			// 实时打印结果
			fmt.Printf("%s %s%-7d %-40s [%6d bytes]\033[0m\n",
				statusSymbol,
				statusColor,
				result.StatusCode,
				result.URL,
				result.ContentLength)
		}

		// 保存结果
		resultFile := filepath.Join(outputPath, "scan_results.json")
		data, err := json.MarshalIndent(results, "", "    ")
		if err != nil {
			return fmt.Errorf("序列化结果失败: %v", err)
		}

		if err := os.WriteFile(resultFile, data, 0644); err != nil {
			return fmt.Errorf("保存结果失败: %v", err)
		}

		// 打印扫描完成摘要
		fmt.Printf("\n%s\n", strings.Repeat("-", 80))
		fmt.Printf("\n[*] 扫描完成! 总共 \033[34m%d\033[0m 个结果\n", len(results))

		// 按状态码分组统计
		statusGroups := make(map[int]int)
		for _, r := range results {
			statusGroups[r.StatusCode]++
		}

		// 打印状态码统计
		for code, count := range statusGroups {
			statusColor := ""
			switch {
			case code >= 200 && code < 300:
				statusColor = "\033[32m" // 绿色
			case code >= 300 && code < 400:
				statusColor = "\033[36m" // 青色
			case code >= 400 && code < 500:
				statusColor = "\033[33m" // 黄色
			case code >= 500:
				statusColor = "\033[31m" // 红色
			}
			fmt.Printf("[*] 状态码 %s%d\033[0m: %d 个\n", statusColor, code, count)
		}

		fmt.Printf("[*] 结果已保存到: \033[34m%s\033[0m\n", resultFile)

		utils.Info("扫描完成",
			zap.String("target", targetURL),
			zap.Int("results", len(results)),
			zap.String("output", resultFile),
		)

		return nil
	},
}

// 初始化扫描命令
func init() {
	rootCmd.AddCommand(scanCmd)

	// 添加扫描命令特有的标志
	scanCmd.Flags().StringP("url", "u", "", "目标URL (必填)")
	scanCmd.Flags().StringP("mode", "m", "fast", "扫描模式 (fast/normal/deep)")
	scanCmd.Flags().BoolP("ai", "a", false, "启用AI增强")
}

// 解析持续时间字符串
func parseDuration(s string) (time.Duration, error) {
	// 如果已经是有效的持续时间格式，直接解析
	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}

	// 尝试添加单位
	if i, err := strconv.Atoi(s); err == nil {
		// 默认单位为毫秒
		return time.Duration(i) * time.Millisecond, nil
	}

	return 0, fmt.Errorf("无法解析持续时间: %s", s)
}

// 解析整数列表
func parseIntList(s string) []int {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))

	for _, part := range parts {
		if i, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
			result = append(result, i)
		}
	}

	return result
}

// 解析速率限制
func parseRateLimit(s string) (int, time.Duration, error) {
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("无效的速率限制格式: %s", s)
	}

	rate, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("无效的速率: %s", parts[0])
	}

	per, err := parseDuration(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("无效的时间单位: %s", parts[1])
	}

	return rate, per, nil
}

// 打印扫描配置
func printScanConfig(s *scanner.Scanner) {
	fmt.Println("扫描配置:")
	fmt.Printf("  目标: %s\n", s.Target)
	fmt.Printf("  并发: %d\n", s.Concurrency)
	fmt.Printf("  超时: %s\n", s.Timeout)
	fmt.Printf("  模式: %s\n", s.Mode)
	fmt.Printf("  技术栈检测: %s\n", s.TechDetect)
	fmt.Printf("  字典大小: %d\n", len(s.Wordlist))
	fmt.Printf("  AI启用: %v\n", s.AIEnabled)
	if s.AIEnabled {
		fmt.Printf("  AI模型: %s\n", s.AIModel)
	}
	fmt.Println()
}

// 打印扫描结果
func printResults(results []scanner.ScanResult, duration time.Duration) {
	fmt.Println("\n扫描结果:")
	fmt.Printf("  总耗时: %s\n", duration)
	fmt.Printf("  发现路径: %d\n\n", len(results))

	// 按状态码分组
	grouped := make(map[int][]scanner.ScanResult)
	for _, result := range results {
		grouped[result.StatusCode] = append(grouped[result.StatusCode], result)
	}

	// 打印结果
	for code, codeResults := range grouped {
		// 选择状态码颜色
		var codeColor *color.Color
		switch {
		case code >= 200 && code < 300:
			codeColor = color.New(color.FgGreen, color.Bold)
		case code >= 300 && code < 400:
			codeColor = color.New(color.FgCyan, color.Bold)
		case code >= 400 && code < 500:
			codeColor = color.New(color.FgYellow, color.Bold)
		case code >= 500:
			codeColor = color.New(color.FgRed, color.Bold)
		default:
			codeColor = color.New(color.FgWhite, color.Bold)
		}

		// 打印状态码组
		codeColor.Printf("[%d] - %d 个结果\n", code, len(codeResults))

		// 打印每个结果
		for _, result := range codeResults {
			fmt.Printf("  %s [%d字节] [%s]\n",
				result.URL,
				result.ContentLength,
				result.ResponseTime,
			)

			// 如果检测到技术栈，显示它
			if len(result.TechStack) > 0 {
				fmt.Printf("    技术栈: %s\n", strings.Join(result.TechStack, ", "))
			}
		}
		fmt.Println()
	}
}

// 打印AI分析结果
func printAIAnalysis(analyses []ai.ScanResultAnalysis) {
	fmt.Println("\nAI分析结果:")
	fmt.Printf("  分析路径: %d\n\n", len(analyses))

	// 按风险级别分组
	grouped := make(map[string][]ai.ScanResultAnalysis)
	for _, analysis := range analyses {
		grouped[analysis.RiskLevel] = append(grouped[analysis.RiskLevel], analysis)
	}

	// 打印高风险结果
	if highRisk, ok := grouped["high"]; ok && len(highRisk) > 0 {
		color.New(color.FgRed, color.Bold).Printf("[高风险] - %d 个结果\n", len(highRisk))
		for _, analysis := range highRisk {
			fmt.Printf("  %s [%d]\n", analysis.OriginalResult.URL, analysis.OriginalResult.StatusCode)
			if len(analysis.Findings) > 0 {
				fmt.Printf("    发现: %s\n", strings.Join(analysis.Findings, ", "))
			}
			if len(analysis.Suggestions) > 0 {
				fmt.Printf("    建议: %s\n", strings.Join(analysis.Suggestions, ", "))
			}
		}
		fmt.Println()
	}

	// 打印中风险结果
	if mediumRisk, ok := grouped["medium"]; ok && len(mediumRisk) > 0 {
		color.New(color.FgYellow, color.Bold).Printf("[中风险] - %d 个结果\n", len(mediumRisk))
		for _, analysis := range mediumRisk {
			fmt.Printf("  %s [%d]\n", analysis.OriginalResult.URL, analysis.OriginalResult.StatusCode)
		}
		fmt.Println()
	}

	// 打印低风险结果数量
	if lowRisk, ok := grouped["low"]; ok && len(lowRisk) > 0 {
		color.New(color.FgGreen).Printf("[低风险] - %d 个结果\n", len(lowRisk))
		fmt.Println()
	}
}

// 从文件读取路径列表
func readPathsFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var paths []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		path := strings.TrimSpace(scanner.Text())
		if path != "" && !strings.HasPrefix(path, "#") {
			paths = append(paths, path)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return paths, nil
}
