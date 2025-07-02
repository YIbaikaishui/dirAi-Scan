package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/diraiscan/internal/pkg/utils"
	"github.com/diraiscan/internal/infra/ai"
	"github.com/diraiscan/internal/domain/model"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	analyzeFile string
	aiModel     string
	aiEndpoint  string
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "分析扫描结果",
	Long:  `使用AI分析扫描结果，识别潜在的安全风险和有价值的发现`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if analyzeFile == "" {
			return fmt.Errorf("请指定要分析的结果文件")
		}

		// 读取扫描结果文件
		data, err := os.ReadFile(analyzeFile)
		if err != nil {
			return fmt.Errorf("读取结果文件失败: %w", err)
		}

		var results []model.Result
		if err := json.Unmarshal(data, &results); err != nil {
			return fmt.Errorf("解析结果文件失败: %w", err)
		}

		utils.Info("开始AI分析", zap.String("file", analyzeFile), zap.Int("results", len(results)))

		// 初始化AI分析器
		analyzer, err := ai.NewAnalyzer(aiEndpoint, aiModel)
		if err != nil {
			return fmt.Errorf("初始化AI分析器失败: %w", err)
		}

		// 执行分析
		analyses, err := analyzer.AnalyzeResults(results)
		if err != nil {
			return fmt.Errorf("AI分析失败: %w", err)
		}

		// 显示分析结果
		printAIAnalysisResults(analyses)

		// 保存分析结果
		outputFile := getAnalysisOutputFile(analyzeFile)
		analysisData, err := json.MarshalIndent(analyses, "", "    ")
		if err != nil {
			return fmt.Errorf("序列化分析结果失败: %w", err)
		}

		if err := os.WriteFile(outputFile, analysisData, 0644); err != nil {
			return fmt.Errorf("保存分析结果失败: %w", err)
		}

		utils.Info("分析完成", zap.String("output", outputFile))
		return nil
	},
}

func init() {
	analyzeCmd.Flags().StringVarP(&analyzeFile, "file", "f", "", "扫描结果文件路径")
	analyzeCmd.Flags().StringVarP(&aiModel, "ai-model", "m", "deepseek-r1", "AI模型名称")
	analyzeCmd.Flags().StringVarP(&aiEndpoint, "ai-endpoint", "e", "http://localhost:11434", "AI服务端点")

	analyzeCmd.MarkFlagRequired("file")
}

// getAnalysisOutputFile 生成分析结果输出文件名
func getAnalysisOutputFile(inputFile string) string {
	ext := filepath.Ext(inputFile)
	base := inputFile[:len(inputFile)-len(ext)]
	return base + "_analysis" + ext
}

// printAIAnalysisResults 显示AI分析结果
func printAIAnalysisResults(analyses []ai.ScanResultAnalysis) {
	if len(analyses) == 0 {
		color.New(color.FgYellow).Println("\n[!] 没有分析结果")
		return
	}

	// 按风险级别分组
	grouped := make(map[string][]ai.ScanResultAnalysis)
	for _, analysis := range analyses {
		grouped[analysis.RiskLevel] = append(grouped[analysis.RiskLevel], analysis)
	}

	color.New(color.FgCyan, color.Bold).Printf("\n[+] AI分析完成，共分析 %d 个结果\n\n", len(analyses))

	// 显示高风险结果
	if highRisk, ok := grouped["high"]; ok && len(highRisk) > 0 {
		color.New(color.FgRed, color.Bold).Printf("🚨 [高风险] - %d 个结果\n", len(highRisk))
		for _, analysis := range highRisk {
			fmt.Printf("  %s [%d]\n", analysis.OriginalResult.URL, analysis.OriginalResult.StatusCode)
			if len(analysis.Findings) > 0 {
				color.New(color.FgRed).Printf("    发现: %s\n", analysis.Findings[0])
			}
			if len(analysis.Suggestions) > 0 {
				color.New(color.FgYellow).Printf("    建议: %s\n", analysis.Suggestions[0])
			}
			color.New(color.FgHiBlack).Printf("    置信度: %.1f%%\n\n", analysis.Confidence*100)
		}
	}

	// 显示中风险结果
	if mediumRisk, ok := grouped["medium"]; ok && len(mediumRisk) > 0 {
		color.New(color.FgYellow, color.Bold).Printf("⚠️  [中风险] - %d 个结果\n", len(mediumRisk))
		for _, analysis := range mediumRisk {
			fmt.Printf("  %s [%d]\n", analysis.OriginalResult.URL, analysis.OriginalResult.StatusCode)
			if len(analysis.Findings) > 0 {
				color.New(color.FgYellow).Printf("    发现: %s\n", analysis.Findings[0])
			}
		}
		fmt.Println()
	}

	// 显示低风险结果
	if lowRisk, ok := grouped["low"]; ok && len(lowRisk) > 0 {
		color.New(color.FgGreen).Printf("ℹ️  [低风险] - %d 个结果\n", len(lowRisk))
		fmt.Println()
	}

	// 显示统计信息
	color.New(color.FgCyan).Printf("📊 风险统计:\n")
	for level, results := range grouped {
		var levelColor *color.Color
		switch level {
		case "high":
			levelColor = color.New(color.FgRed, color.Bold)
		case "medium":
			levelColor = color.New(color.FgYellow, color.Bold)
		case "low":
			levelColor = color.New(color.FgGreen)
		default:
			levelColor = color.New(color.FgWhite)
		}
		levelColor.Printf("  %s: %d 个\n", level, len(results))
	}
}
