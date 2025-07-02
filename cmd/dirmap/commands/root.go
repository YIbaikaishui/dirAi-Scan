package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dirmap",
	Short: "DirAI-Scan - 智能化Web目录扫描工具",
	Long: `DirAI-Scan 是一个基于Go语言开发的高性能、智能化Web目录扫描工具。

特性:
- 高并发扫描引擎
- AI智能分析
- 多模式扫描
- 技术栈识别
- 灵活的配置系统`,
	Version: "1.0.0",
}

// Execute 执行根命令
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// 添加全局标志
	rootCmd.PersistentFlags().StringP("config", "c", "", "配置文件路径")
	rootCmd.PersistentFlags().StringP("log-level", "l", "info", "日志级别 (debug, info, warn, error)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "详细输出")

	// 添加子命令
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(analyzeCmd)
}

func init() {
	// 设置版本模板
	rootCmd.SetVersionTemplate(`{{printf "%s version %s\n" .Name .Version}}`)

	// 自定义帮助命令
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "help [command]",
		Short:  "获取任何命令的帮助信息",
		Long:   `获取任何命令的帮助信息`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Root().Help()
				return
			}
			if err := cmd.Root().Help(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
		},
	})
}