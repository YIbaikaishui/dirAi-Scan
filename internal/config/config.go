package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config 应用程序配置结构
type Config struct {
	// 扫描配置
	Scan struct {
		Threads    int    `json:"threads"`     // 并发线程数
		Timeout    int    `json:"timeout"`     // 超时时间(秒)
		UserAgent  string `json:"user_agent"`  // 用户代理
		Proxy      string `json:"proxy"`       // 代理设置
		RetryTimes int    `json:"retry_times"` // 重试次数
		RetryDelay int    `json:"retry_delay"` // 重试延迟(秒)
	} `json:"scan"`

	// AI配置
	AI struct {
		Enabled  bool   `json:"enabled"`  // 是否启用AI
		Model    string `json:"model"`    // AI模型
		APIKey   string `json:"api_key"`  // API密钥(留空本地模型)
		Endpoint string `json:"endpoint"` // API端点(本地为http://localhost:11434)
		Local    bool   `json:"local"`    // 是否为本地模型
	} `json:"ai"`

	// 日志配置
	LogDir    string `json:"log_dir"`    // 日志目录
	LogLevel  string `json:"log_level"`  // 日志级别
	LogFormat string `json:"log_format"` // 日志格式

	// 输出配置
	Output struct {
		Format  string `json:"format"`  // 输出格式(json/csv/html)
		Path    string `json:"path"`    // 输出路径
		Verbose bool   `json:"verbose"` // 是否详细输出
	} `json:"output"`
}

var (
	config *Config
)

// LoadConfig 从文件加载配置
func LoadConfig(configPath string) error {
	if configPath == "" {
		configPath = "config.json"
	}

	// 如果配置文件不存在，创建默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config = &Config{}
		setDefaultConfig(config)
		// 优先从环境变量加载AI配置
		loadAIEnv(config)
		return SaveConfig(configPath)
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	config = &Config{}
	if err := json.Unmarshal(data, config); err != nil {
		return err
	}

	// 优先从环境变量加载AI配置
	loadAIEnv(config)

	return nil
}

// SaveConfig 保存配置到文件
func SaveConfig(configPath string) error {
	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}

	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// GetConfig 获取当前配置
func GetConfig() *Config {
	if config == nil {
		config = &Config{}
		setDefaultConfig(config)
	}
	return config
}

// setDefaultConfig 设置默认配置
func setDefaultConfig(cfg *Config) {
	// 扫描默认配置
	cfg.Scan.Threads = 100
	cfg.Scan.Timeout = 30
	cfg.Scan.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	cfg.Scan.RetryTimes = 3
	cfg.Scan.RetryDelay = 5

	// AI默认配置
	cfg.AI.Enabled = false
	cfg.AI.Model = "qwen3:4b"
	cfg.AI.APIKey = ""
	cfg.AI.Endpoint = "http://localhost:11434"
	cfg.AI.Local = true
	// 日志默认配置
	cfg.LogDir = "logs"
	cfg.LogLevel = "info"
	cfg.LogFormat = "json"

	// 输出默认配置
	cfg.Output.Format = "json"
	cfg.Output.Path = "results"
	cfg.Output.Verbose = false
}

// loadAIEnv 优先从环境变量加载AI配置
func loadAIEnv(cfg *Config) {
	if v := os.Getenv("AI_KEY"); v != "" {
		cfg.AI.APIKey = v
	}
	if v := os.Getenv("AI_ENDPOINT"); v != "" {
		cfg.AI.Endpoint = v
	}
	if v := os.Getenv("AI_MODEL"); v != "" {
		cfg.AI.Model = v
	}
	if v := os.Getenv("AI_LOCAL"); v != "" {
		cfg.AI.Local = v == "true" || v == "1"
	}
	if v := os.Getenv("AI_ENABLED"); v != "" {
		cfg.AI.Enabled = v == "true" || v == "1"
	}
}
