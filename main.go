package main

import (
	"fmt"
	"os"

	"github.com/diraiscan/cmd"
	"github.com/diraiscan/internal/config"
	"github.com/diraiscan/internal/utils"

	"github.com/joho/godotenv"
)

func main() {
	// 优先加载.env环境变量
	_ = godotenv.Load()

	// 初始化配置
	if err := config.LoadConfig(""); err != nil {
		fmt.Printf("警告: 无法加载配置文件: %v\n", err)
	}

	// 初始化日志
	logConfig := config.GetConfig()
	if err := utils.InitLogger(logConfig.LogDir, logConfig.LogLevel); err != nil {
		fmt.Printf("警告: 无法初始化日志: %v\n", err)
	}

	// 执行命令
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
