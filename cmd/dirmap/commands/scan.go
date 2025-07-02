package commands

import (
	"sync"

	"github.com/diraiscan/internal/pkg/utils"
	"github.com/diraiscan/internal/infra/ai"
	"github.com/diraiscan/internal/domain/model"

	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	scanTarget  string
	outputFile  string
	commonFile  string
	verboseMode bool
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "开始扫描目标",
	RunE: func(cmd *cobra.Command, args []string) (err error) { // Named return for panic handling
		// 获取verbose标志
		verboseMode, _ = cmd.Flags().GetBool("verbose")

		// Defer a panic handler for the entire RunE function
		defer func() {
			if r := recover(); r != nil {
				utils.Error("RunE 函数中发生 panic", zap.Any("panic", r))
				// Ensure an error is returned if a panic occurred
				if err == nil {
					err = fmt.Errorf("panic occurred: %v", r)
				}
			}
		}()

		// cfg := config.GetConfig() // 移除未使用的 cfg 变量

		var paths []string
		dataDir := "./data"
		files, err := os.ReadDir(dataDir)
		if err != nil {
			utils.Error("读取data目录失败", zap.Error(err))
			return err
		}

		for _, file := range files {
			if !file.IsDir() {
				filePath := filepath.Join(dataDir, file.Name())
				content, err := os.ReadFile(filePath)
				if err != nil {
					utils.Error("读取文件内容失败", zap.Error(err), zap.String("file", filePath))
					continue
				}
				lines := strings.Split(string(content), "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line != "" && !strings.HasPrefix(line, "#") {
						paths = append(paths, line)
					}
				}
			}
		}

		if len(paths) == 0 {
			utils.Warn("未从data目录读取到任何扫描路径，将使用默认路径")
			paths = []string{
				"/.env",
				"/admin",
				"/login",
				"/robots.txt",
				"/sitemap.xml",
			}
		}

		var targetURLs []string
		if scanTarget != "" {
			targetURLs = append(targetURLs, scanTarget)
		} else if commonFile != "" {
			file, err := os.Open(commonFile)
			if err != nil {
				return fmt.Errorf("无法打开common.txt文件: %w", err)
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" {
					targetURLs = append(targetURLs, line)
				}
			}
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("读取common.txt文件时发生错误: %w", err)
			}
		} else {
			return fmt.Errorf("请提供目标URL或指定common.txt文件路径")
		}

		if len(targetURLs) == 0 {
			return fmt.Errorf("没有找到任何扫描目标")
		}

		for i, url := range targetURLs {
			if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
				targetURLs[i] = "http://" + url
			}
		}

		if outputFile == "" {
			outputFile = "results"
		}

		if err := os.MkdirAll(outputFile, 0755); err != nil {
			return fmt.Errorf("创建输出目录失败: %v", err)
		}

		// 从cobra命令中获取配置参数
		concurrency, _ := cmd.Flags().GetInt("threads")
		timeoutStr, _ := cmd.Flags().GetString("timeout")
		timeout, err := parseDuration(timeoutStr)
		if err != nil {
			utils.Error("无效的超时时间格式", zap.String("timeout", timeoutStr), zap.Error(err))
			// 使用默认超时时间
			timeout = 5 * time.Second
		}
		statusCodesStr, _ := cmd.Flags().GetString("status-codes")
		statusCodes := parseIntList(statusCodesStr)

		aiEnabled, _ := cmd.Flags().GetBool("ai-enable")
		aiModel, _ := cmd.Flags().GetString("ai-model")
		aiEndpoint, _ := cmd.Flags().GetString("ai-endpoint")

		for _, targetURL := range targetURLs {
			options := []func(*model.Scanner){
				model.WithWordlist(paths),
				model.WithConcurrency(concurrency),
				model.WithTimeout(timeout),
				model.WithAI(aiEnabled, aiModel, aiEndpoint),
			}
			if len(statusCodes) > 0 {
				options = append(options, model.WithStatusCodes(statusCodes))
			}

			engine := model.NewScanner(targetURL, options...)

			// 显示扫描开始横幅
			printScanBanner(targetURL, len(paths), concurrency, timeout)

			utils.Info("开始扫描", zap.String("target", targetURL))

			// 创建一个回调函数，用于实时显示扫描结果
			var liveResults []model.Result
			var resultMutex sync.Mutex

			// 定义回调函数，实时处理和显示扫描结果
			resultCallback := func(result model.Result) {
				resultMutex.Lock()
				liveResults = append(liveResults, result)
				resultMutex.Unlock()

				// 实时显示每个发现的结果
				printSingleResult(result)
			}

			// 使用带回调函数的扫描方法
			results := engine.StartWithCallback(resultCallback)

			// 扫描完成后，使用完整结果集
			scanResults := results
			// 显示扫描统计信息
			printScanSummary(scanResults, targetURL)

			// 按状态码分组显示结果
			printResultsByStatus(scanResults)

			// 错误处理逻辑可以保留，如果 Scanner 内部有错误通道的话
			// for err := range engine.GetErrors() { // 假设 Scanner 有 GetErrors 方法
			// 	utils.Error("扫描错误", zap.Error(err), zap.String("target", targetURL))
			// }

			resultFile := filepath.Join(outputFile, fmt.Sprintf("scan_results_%s.json", strings.ReplaceAll(strings.ReplaceAll(targetURL, "http://", ""), "https://", "")))
			data, err := json.MarshalIndent(scanResults, "", "    ")
			if err != nil {
				utils.Error("序列化结果失败", zap.Error(err), zap.String("target", targetURL))
				// 如果在 for targetURL 循环中，这里应该是 continue
				// 但由于这是在 RunE 函数的顶层，所以应该是 return err
				return err
			}

			if err := os.WriteFile(resultFile, data, 0644); err != nil {
				utils.Error("保存结果失败", zap.Error(err), zap.String("target", targetURL))
				// 同上，这里应该是 return err
				return err
			}

			utils.Info("扫描完成",
				zap.String("target", targetURL),
				zap.Int("results", len(scanResults)),
				zap.String("output", resultFile),
			)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().StringVarP(&scanTarget, "url", "u", "", "目标URL (e.g., http://example.com)")
	scanCmd.Flags().StringVarP(&outputFile, "output", "o", "", "输出文件路径")
	scanCmd.Flags().StringVarP(&commonFile, "common", "C", "", "从common.txt文件读取路径")

	scanCmd.Flags().StringP("mode", "m", "fast", "扫描模式 (fast/normal/deep)")
	scanCmd.Flags().BoolP("ai", "a", false, "启用AI增强")
}

func parseDuration(s string) (time.Duration, error) {
	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}

	if i, err := strconv.Atoi(s); err == nil {
		return time.Duration(i) * time.Millisecond, nil
	}

	return 0, fmt.Errorf("无法解析持续时间: %s", s)
}

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

func printScanConfig(s *model.Scanner) {
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

func printResults(results []model.Result, duration time.Duration) {
	fmt.Println("\n扫描结果:")
	fmt.Printf("  总耗时: %s\n", duration)
	fmt.Printf("  发现路径: %d\n\n", len(results))

	grouped := make(map[int][]model.Result)
	for _, result := range results {
		grouped[result.StatusCode] = append(grouped[result.StatusCode], result)
	}

	for code, codeResults := range grouped {
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

		codeColor.Printf("[%d] - %d 个结果\n", code, len(codeResults))

		for _, result := range codeResults {
			fmt.Printf("  %s [%d字节] [%s]\n",
				result.URL,
				result.ContentLength,
				result.ResponseTime,
			)

			if len(result.TechStack) > 0 {
				fmt.Printf("    技术栈: %s\n", strings.Join(result.TechStack, ", "))
			}
		}
		fmt.Println()
	}
}

func printAIAnalysis(analyses []ai.ScanResultAnalysis) {
	fmt.Println("\nAI分析结果:")
	fmt.Printf("  分析路径: %d\n\n", len(analyses))

	grouped := make(map[string][]ai.ScanResultAnalysis)
	for _, analysis := range analyses {
		grouped[analysis.RiskLevel] = append(grouped[analysis.RiskLevel], analysis)
	}

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

	if mediumRisk, ok := grouped["medium"]; ok && len(mediumRisk) > 0 {
		color.New(color.FgYellow, color.Bold).Printf("[中风险] - %d 个结果\n", len(mediumRisk))
		for _, analysis := range mediumRisk {
			fmt.Printf("  %s [%d]\n", analysis.OriginalResult.URL, analysis.OriginalResult.StatusCode)
		}
		fmt.Println()
	}

	if lowRisk, ok := grouped["low"]; ok && len(lowRisk) > 0 {
		color.New(color.FgGreen).Printf("[低风险] - %d 个结果\n", len(lowRisk))
		fmt.Println()
	}
}

func printSingleResult(result model.Result) {
	// 根据状态码设置不同的颜色和图标
	var statusColor *color.Color
	var statusIcon string

	switch {
	case result.StatusCode >= 200 && result.StatusCode < 300:
		// 绿色 - 成功找到的资源
		statusColor = color.New(color.FgGreen, color.Bold)
		statusIcon = "✓"
	case result.StatusCode >= 300 && result.StatusCode < 400:
		// 青色 - 重定向
		statusColor = color.New(color.FgCyan, color.Bold)
		statusIcon = "→"
	case result.StatusCode == 403:
		// 黄色 - 禁止访问（可能存在但无权限）
		statusColor = color.New(color.FgYellow, color.Bold)
		statusIcon = "⚠"
	case result.StatusCode >= 400 && result.StatusCode < 500:
		// 橙色/黄色 - 客户端错误
		statusColor = color.New(color.FgYellow)
		statusIcon = "!"
	case result.StatusCode >= 500:
		// 红色 - 服务器错误
		statusColor = color.New(color.FgRed, color.Bold)
		statusIcon = "✗"
	default:
		// 白色 - 其他状态
		statusColor = color.New(color.FgWhite)
		statusIcon = "?"
	}

	// URL路径颜色 - 蓝色显示扫描的路径
	urlColor := color.New(color.FgBlue, color.Bold)

	// 内容长度颜色 - 根据大小设置不同颜色
	var sizeColor *color.Color
	if result.ContentLength > 10000 {
		sizeColor = color.New(color.FgMagenta, color.Bold) // 大文件用粗体紫色
	} else if result.ContentLength > 1000 {
		sizeColor = color.New(color.FgMagenta) // 中等文件用紫色
	} else {
		sizeColor = color.New(color.FgHiBlack) // 小文件用灰色
	}

	// 响应时间颜色 - 根据响应时间设置颜色
	var timeColor *color.Color
	if result.ResponseTime > 2*time.Second {
		timeColor = color.New(color.FgRed) // 慢响应用红色
	} else if result.ResponseTime > 1*time.Second {
		timeColor = color.New(color.FgYellow) // 中等响应用黄色
	} else {
		timeColor = color.New(color.FgGreen) // 快响应用绿色
	}

	// 技术栈颜色
	techColor := color.New(color.FgHiYellow, color.Bold)

	// 输出格式化结果 - 类似 dirsearch 的格式
	statusColor.Printf("%s [%d] ", statusIcon, result.StatusCode)
	urlColor.Printf("%s", result.URL)

	// 显示文件大小信息
	fmt.Printf(" (")
	sizeColor.Printf("%s", formatFileSize(result.ContentLength))
	fmt.Printf(") ")

	// 显示响应时间
	fmt.Printf("[")
	timeColor.Printf("%s", result.ResponseTime.Truncate(time.Millisecond))
	fmt.Printf("]")

	// 显示技术栈信息
	if len(result.TechStack) > 0 {
		fmt.Printf(" [")
		techColor.Printf("Tech: %s", strings.Join(result.TechStack, ", "))
		fmt.Printf("]")
	}

	// 显示重要的响应头信息
	if server, ok := result.Headers["Server"]; ok {
		fmt.Printf(" [")
		color.New(color.FgCyan).Printf("Server: %s", server)
		fmt.Printf("]")
	}

	fmt.Println()

	// 详细模式下显示更多信息
	if verboseMode {
		// 显示重要的响应头信息
		headerColor := color.New(color.FgHiBlack)
		importantHeaders := []string{"Content-Type", "Location", "Set-Cookie", "X-Powered-By", "X-Frame-Options"}

		for _, header := range importantHeaders {
			if value, ok := result.Headers[header]; ok {
				headerColor.Printf("    %s: %s\n", header, value)
			}
		}

		// 显示响应体摘要（如果有）
		if result.Body != "" {
			body := result.Body
			if len(body) > 150 {
				body = body[:150] + "..."
			}
			// 移除换行符，使输出更紧凑
			body = strings.ReplaceAll(body, "\n", " ")
			body = strings.ReplaceAll(body, "\r", "")
			headerColor.Printf("    Preview: %s\n", body)
		}

		// 显示空行分隔
		fmt.Println()
	}
}

// formatFileSize 格式化文件大小显示
func formatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%dB", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1fKB", float64(size)/1024)
	} else {
		return fmt.Sprintf("%.1fMB", float64(size)/(1024*1024))
	}
}

// printScanSummary 显示扫描摘要信息
func printScanSummary(results []model.Result, target string) {
	if len(results) == 0 {
		color.New(color.FgRed).Printf("\n[!] 未发现任何有效路径: %s\n\n", target)
		return
	}

	// 统计不同状态码的数量
	statusStats := make(map[int]int)
	for _, result := range results {
		statusStats[result.StatusCode]++
	}

	// 显示扫描摘要
	color.New(color.FgCyan, color.Bold).Printf("\n[+] 扫描完成: %s\n", target)
	color.New(color.FgWhite).Printf("[+] 发现 %d 个有效路径\n\n", len(results))

	// 显示状态码统计
	color.New(color.FgHiBlue, color.Bold).Println("状态码统计:")
	for code, count := range statusStats {
		var statusColor *color.Color
		var statusDesc string

		switch {
		case code >= 200 && code < 300:
			statusColor = color.New(color.FgGreen, color.Bold)
			statusDesc = "成功"
		case code >= 300 && code < 400:
			statusColor = color.New(color.FgCyan, color.Bold)
			statusDesc = "重定向"
		case code == 403:
			statusColor = color.New(color.FgYellow, color.Bold)
			statusDesc = "禁止访问"
		case code >= 400 && code < 500:
			statusColor = color.New(color.FgYellow)
			statusDesc = "客户端错误"
		case code >= 500:
			statusColor = color.New(color.FgRed, color.Bold)
			statusDesc = "服务器错误"
		default:
			statusColor = color.New(color.FgWhite)
			statusDesc = "其他"
		}

		statusColor.Printf("  [%d] %s: %d 个\n", code, statusDesc, count)
	}
	fmt.Println()
}

// printResultsByStatus 按状态码分组显示结果
func printResultsByStatus(results []model.Result) {
	if len(results) == 0 {
		return
	}

	// 按状态码分组
	grouped := make(map[int][]model.Result)
	for _, result := range results {
		grouped[result.StatusCode] = append(grouped[result.StatusCode], result)
	}

	// 按优先级顺序显示结果
	priority := []int{200, 201, 202, 204, 301, 302, 303, 307, 308, 403, 401, 405, 500, 502, 503}

	// 首先显示高优先级的状态码
	for _, code := range priority {
		if results, exists := grouped[code]; exists {
			printStatusGroup(code, results)
			delete(grouped, code) // 从map中移除已显示的
		}
	}

	// 显示剩余的状态码
	for code, results := range grouped {
		printStatusGroup(code, results)
	}
}

// printStatusGroup 显示特定状态码的结果组
func printStatusGroup(statusCode int, results []model.Result) {
	if len(results) == 0 {
		return
	}

	var headerColor *color.Color
	var statusDesc string
	var statusIcon string

	switch {
	case statusCode >= 200 && statusCode < 300:
		headerColor = color.New(color.FgGreen, color.Bold)
		statusDesc = "成功响应"
		statusIcon = "✓"
	case statusCode >= 300 && statusCode < 400:
		headerColor = color.New(color.FgCyan, color.Bold)
		statusDesc = "重定向"
		statusIcon = "→"
	case statusCode == 403:
		headerColor = color.New(color.FgYellow, color.Bold)
		statusDesc = "禁止访问"
		statusIcon = "⚠"
	case statusCode >= 400 && statusCode < 500:
		headerColor = color.New(color.FgYellow)
		statusDesc = "客户端错误"
		statusIcon = "!"
	case statusCode >= 500:
		headerColor = color.New(color.FgRed, color.Bold)
		statusDesc = "服务器错误"
		statusIcon = "✗"
	default:
		headerColor = color.New(color.FgWhite)
		statusDesc = "其他响应"
		statusIcon = "?"
	}

	// 显示状态码组标题
	headerColor.Printf("%s [%d] %s (%d 个结果):\n", statusIcon, statusCode, statusDesc, len(results))

	// 显示该状态码下的所有结果
	for _, result := range results {
		printSingleResult(result)
	}

	fmt.Println() // 组之间的分隔
}

// printScanBanner 显示扫描开始横幅
func printScanBanner(target string, wordlistSize, concurrency int, timeout time.Duration) {
	// 显示工具横幅
	color.New(color.FgHiCyan, color.Bold).Println("\n╔══════════════════════════════════════════════════════════════════════════════╗")
	color.New(color.FgHiCyan, color.Bold).Println("║                              DirAI-Scan v1.0                                ║")
	color.New(color.FgHiCyan, color.Bold).Println("║                        AI-Enhanced Directory Scanner                        ║")
	color.New(color.FgHiCyan, color.Bold).Println("╚══════════════════════════════════════════════════════════════════════════════╝")

	// 显示扫描配置
	color.New(color.FgHiWhite, color.Bold).Println("\n[扫描配置]")
	color.New(color.FgWhite).Printf("目标URL:     ")
	color.New(color.FgHiBlue, color.Bold).Printf("%s\n", target)

	color.New(color.FgWhite).Printf("字典大小:    ")
	color.New(color.FgHiGreen, color.Bold).Printf("%d 个路径\n", wordlistSize)

	color.New(color.FgWhite).Printf("并发线程:    ")
	color.New(color.FgHiYellow, color.Bold).Printf("%d\n", concurrency)

	color.New(color.FgWhite).Printf("请求超时:    ")
	color.New(color.FgHiMagenta, color.Bold).Printf("%s\n", timeout)

	// 显示扫描状态说明
	color.New(color.FgHiWhite, color.Bold).Println("\n[状态码说明]")
	color.New(color.FgGreen, color.Bold).Printf("✓ 2xx ")
	color.New(color.FgWhite).Printf("成功响应  ")
	color.New(color.FgCyan, color.Bold).Printf("→ 3xx ")
	color.New(color.FgWhite).Printf("重定向  ")
	color.New(color.FgYellow, color.Bold).Printf("⚠ 403 ")
	color.New(color.FgWhite).Printf("禁止访问  ")
	color.New(color.FgYellow).Printf("! 4xx ")
	color.New(color.FgWhite).Printf("客户端错误  ")
	color.New(color.FgRed, color.Bold).Printf("✗ 5xx ")
	color.New(color.FgWhite).Printf("服务器错误\n")

	color.New(color.FgHiCyan, color.Bold).Println("\n" + strings.Repeat("=", 80))
	color.New(color.FgHiWhite, color.Bold).Printf("开始扫描 %s\n", target)
	color.New(color.FgHiCyan, color.Bold).Println(strings.Repeat("=", 80))
	fmt.Println()
}
