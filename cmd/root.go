package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
	rootCmd = &cobra.Command{
		Use:   "diraiscan",
		Short: "DirAI-Scan - 下一代高并发智能Web扫描器",
		Long: `DirAI-Scan 是一个基于Go语言的高并发Web扫描器，具有以下特性：
- 高并发扫描引擎
- AI智能决策系统
- 多模式扫描技术
- 现代框架识别能力`,
		Run: func(cmd *cobra.Command, args []string) {
			// 如果没有子命令，显示帮助信息
			cmd.Help()
		},
	}
)

// Execute 执行根命令
func Execute() error {
	// 显示欢迎信息
	printBanner()

	// 执行命令
	return rootCmd.Execute()
}

// 初始化命令行参数
func init() {
	// 添加全局标志
	rootCmd.PersistentFlags().StringP("config", "c", "", "配置文件路径")
	rootCmd.PersistentFlags().StringP("output", "o", "", "输出文件路径")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "详细输出模式")

	// 基础扫描参数
	rootCmd.PersistentFlags().StringP("url", "u", "", "目标URL (必填)")
	rootCmd.PersistentFlags().StringP("wordlist", "w", "", "自定义字典路径（默认使用内置字典）")
	rootCmd.PersistentFlags().StringP("status-codes", "", "200,403,500", "仅显示指定状态码的结果")
	rootCmd.PersistentFlags().IntP("threads", "t", 50, "并发线程数")
	rootCmd.PersistentFlags().StringP("timeout", "", "5s", "请求超时时间")

	// 扫描模式
	rootCmd.PersistentFlags().StringP("mode", "m", "fast", "扫描模式: fast/deep/stealth")
	rootCmd.PersistentFlags().StringP("tech-detect", "", "quick", "技术栈识别深度: quick/full")

	// AI 增强参数
	rootCmd.PersistentFlags().BoolP("ai-enable", "", false, "启用AI分析（需Ollama服务）")
	rootCmd.PersistentFlags().StringP("ai-model", "", "deepseek-r1", "指定AI模型")
	rootCmd.PersistentFlags().StringP("ai-filter", "", "", "仅输出AI标记的特定结果")
	rootCmd.PersistentFlags().StringP("ai-endpoint", "", "http://localhost:11434", "Ollama服务地址")

	// 性能优化
	rootCmd.PersistentFlags().StringP("rate-limit", "", "", "速率限制（如100/5s）")
	rootCmd.PersistentFlags().StringP("proxy", "p", "", "代理服务器（支持HTTP/SOCKS5）")
	rootCmd.PersistentFlags().IntP("retries", "r", 3, "失败请求重试次数")
	rootCmd.PersistentFlags().StringP("delay", "d", "", "请求间固定延迟")

	// 高级功能
	rootCmd.PersistentFlags().StringP("distribute", "", "", "分布式模式: controller/worker")
	rootCmd.PersistentFlags().StringP("plugins", "", "", "加载插件（逗号分隔）")
	rootCmd.PersistentFlags().StringP("cookie", "", "", "自定义Cookie")
	rootCmd.PersistentFlags().StringP("header", "H", "", "自定义HTTP头")

	// 过滤与排除
	rootCmd.PersistentFlags().StringP("exclude-status", "", "404", "排除指定状态码的结果")
	rootCmd.PersistentFlags().IntP("exclude-size", "", 0, "排除特定大小的响应（字节）")
	rootCmd.PersistentFlags().StringP("match-regex", "", "", "仅保留匹配正则的路径")

	// 版本信息
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("DirAI-Scan 版本 %s\n", version)
		},
	})
}

// 打印欢迎横幅
func printBanner() {
	banner := `
██████╗ ██╗██████╗  █████╗ ██╗   ███████╗ ██████╗ █████╗ ███╗   ██╗
██╔══██╗██║██╔══██╗██╔══██╗██║   ██╔════╝██╔════╝██╔══██╗████╗  ██║
██║  ██║██║██████╔╝███████║██║   ███████╗██║     ███████║██╔██╗ ██║
██║  ██║██║██╔══██╗██╔══██║██║   ╚════██║██║     ██╔══██║██║╚██╗██║
██████╔╝██║██║  ██║██║  ██║██║   ███████║╚██████╗██║  ██║██║ ╚████║
╚═════╝ ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝   ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝
                                                         v%s
`
	colorBanner := color.New(color.FgCyan).Sprintf(banner, version)
	fmt.Println(colorBanner)
	color.New(color.FgHiWhite).Println("下一代高并发智能Web扫描器 | 基于Go语言与AI增强")
	fmt.Println()
}
