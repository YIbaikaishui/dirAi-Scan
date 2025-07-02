package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/diraiscan/cmd/dirmap/commands"
	"github.com/diraiscan/internal/pkg/config"
	"github.com/diraiscan/internal/pkg/utils"

	"github.com/joho/godotenv"
)

func main() {
	rand.Seed(time.Now().UnixNano()) // 在程序启动时调用一次 rand.Seed
	_ = godotenv.Load()

	if err := config.LoadConfig(""); err != nil {
		fmt.Printf("警告: 无法加载配置文件: %v\n", err)
	}

	logConfig := config.GetConfig()
	if err := utils.InitLogger(logConfig.LogDir, logConfig.LogLevel); err != nil {
		fmt.Printf("警告: 无法初始化日志: %v\n", err)
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Main 中捕获到 Panic: %v\n", r)
			// 可以在这里添加更详细的堆栈跟踪打印
			os.Exit(2) // 使用与 panic 相同的退出码
		}
	}()

	if err := commands.Execute(); err != nil {
		fmt.Printf("commands.Execute() 返回错误: %v\n", err)
		os.Exit(1)
	}
}
