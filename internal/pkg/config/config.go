package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	App          AppConfig          `mapstructure:"app"`
	Logging      LoggingConfig      `mapstructure:"logging"`
	Scan         ScanConfig         `mapstructure:"scan"`
	AI           AIConfig           `mapstructure:"ai"`
	Network      NetworkConfig      `mapstructure:"network"`
	Output       OutputConfig       `mapstructure:"output"`
	Plugins      PluginsConfig      `mapstructure:"plugins"`
	Dictionaries DictionariesConfig `mapstructure:"dictionaries"`
	Security     SecurityConfig     `mapstructure:"security"`
	Distributed  DistributedConfig  `mapstructure:"distributed"`
	Monitoring   MonitoringConfig   `mapstructure:"monitoring"`
}

// AppConfig 应用基础配置
type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	Debug   bool   `mapstructure:"debug"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Dir        string `mapstructure:"dir"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// ScanConfig 扫描配置
type ScanConfig struct {
	Concurrency    int                    `mapstructure:"concurrency"`
	Timeout        string                 `mapstructure:"timeout"`
	Retries        int                    `mapstructure:"retries"`
	Delay          string                 `mapstructure:"delay"`
	StatusCodes    StatusCodesConfig      `mapstructure:"status_codes"`
	ResponseSize   ResponseSizeConfig     `mapstructure:"response_size"`
	UserAgents     []string               `mapstructure:"user_agents"`
}

// StatusCodesConfig 状态码配置
type StatusCodesConfig struct {
	Include []int `mapstructure:"include"`
	Exclude []int `mapstructure:"exclude"`
}

// ResponseSizeConfig 响应大小配置
type ResponseSizeConfig struct {
	Min          int64   `mapstructure:"min"`
	Max          int64   `mapstructure:"max"`
	ExcludeSizes []int64 `mapstructure:"exclude_sizes"`
}

// AIConfig AI配置
type AIConfig struct {
	Enabled    bool           `mapstructure:"enabled"`
	Endpoint   string         `mapstructure:"endpoint"`
	Model      string         `mapstructure:"model"`
	Timeout    string         `mapstructure:"timeout"`
	MaxRetries int            `mapstructure:"max_retries"`
	Analysis   AnalysisConfig `mapstructure:"analysis"`
}

// AnalysisConfig AI分析配置
type AnalysisConfig struct {
	ConfidenceThreshold float64  `mapstructure:"confidence_threshold"`
	MaxFindings         int      `mapstructure:"max_findings"`
	RiskLevels          []string `mapstructure:"risk_levels"`
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	Proxy     ProxyConfig     `mapstructure:"proxy"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	DNS       DNSConfig       `mapstructure:"dns"`
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	URL      string `mapstructure:"url"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerSecond int  `mapstructure:"requests_per_second"`
	Burst             int  `mapstructure:"burst"`
}

// DNSConfig DNS配置
type DNSConfig struct {
	Servers []string `mapstructure:"servers"`
	Timeout string   `mapstructure:"timeout"`
}

// OutputConfig 输出配置
type OutputConfig struct {
	Format           string       `mapstructure:"format"`
	Directory        string       `mapstructure:"directory"`
	FilenameTemplate string       `mapstructure:"filename_template"`
	Report           ReportConfig `mapstructure:"report"`
}

// ReportConfig 报告配置
type ReportConfig struct {
	Enabled              bool   `mapstructure:"enabled"`
	Template             string `mapstructure:"template"`
	IncludeScreenshots   bool   `mapstructure:"include_screenshots"`
	IncludeRawResponses  bool   `mapstructure:"include_raw_responses"`
}

// PluginsConfig 插件配置
type PluginsConfig struct {
	Enabled   bool     `mapstructure:"enabled"`
	Directory string   `mapstructure:"directory"`
	AutoLoad  []string `mapstructure:"auto_load"`
}

// DictionariesConfig 字典配置
type DictionariesConfig struct {
	Directory string       `mapstructure:"directory"`
	Files     []string     `mapstructure:"files"`
	Custom    CustomConfig `mapstructure:"custom"`
}

// CustomConfig 自定义字典配置
type CustomConfig struct {
	Enabled bool     `mapstructure:"enabled"`
	Paths   []string `mapstructure:"paths"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	Headers map[string]string `mapstructure:"headers"`
	Cookies CookiesConfig     `mapstructure:"cookies"`
	TLS     TLSConfig         `mapstructure:"tls"`
}

// CookiesConfig Cookie配置
type CookiesConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	JarFile string `mapstructure:"jar_file"`
}

// TLSConfig TLS配置
type TLSConfig struct {
	Verify     bool   `mapstructure:"verify"`
	MinVersion string `mapstructure:"min_version"`
	MaxVersion string `mapstructure:"max_version"`
}

// DistributedConfig 分布式配置
type DistributedConfig struct {
	Enabled    bool             `mapstructure:"enabled"`
	Mode       string           `mapstructure:"mode"`
	Controller ControllerConfig `mapstructure:"controller"`
	Worker     WorkerConfig     `mapstructure:"worker"`
}

// ControllerConfig 控制器配置
type ControllerConfig struct {
	BindAddress string `mapstructure:"bind_address"`
	AuthToken   string `mapstructure:"auth_token"`
}

// WorkerConfig 工作节点配置
type WorkerConfig struct {
	ControllerURL string `mapstructure:"controller_url"`
	AuthToken     string `mapstructure:"auth_token"`
	MaxTasks      int    `mapstructure:"max_tasks"`
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	Enabled           bool             `mapstructure:"enabled"`
	MetricsPort       int              `mapstructure:"metrics_port"`
	HealthCheckPort   int              `mapstructure:"health_check_port"`
	Prometheus        PrometheusConfig `mapstructure:"prometheus"`
}

// PrometheusConfig Prometheus配置
type PrometheusConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Endpoint string `mapstructure:"endpoint"`
}

// Manager 配置管理器
type Manager struct {
	config *Config
	viper  *viper.Viper
}

// NewManager 创建配置管理器
func NewManager() *Manager {
	return &Manager{
		viper: viper.New(),
	}
}

// Load 加载配置文件
func (m *Manager) Load(configPath string) error {
	// 设置配置文件路径
	if configPath != "" {
		m.viper.SetConfigFile(configPath)
	} else {
		// 默认配置文件搜索路径
		m.viper.SetConfigName("dirmap")
		m.viper.SetConfigType("yaml")
		m.viper.AddConfigPath(".")
		m.viper.AddConfigPath("./configs")
		m.viper.AddConfigPath("$HOME/.diraiscan")
		m.viper.AddConfigPath("/etc/diraiscan")
	}

	// 设置环境变量前缀
	m.viper.SetEnvPrefix("DIRAISCAN")
	m.viper.AutomaticEnv()

	// 设置默认值
	m.setDefaults()

	// 读取配置文件
	if err := m.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件不存在，使用默认配置
			fmt.Printf("Warning: Config file not found, using defaults\n")
		} else {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// 解析配置到结构体
	m.config = &Config{}
	if err := m.viper.Unmarshal(m.config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 验证配置
	if err := m.validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	return nil
}

// setDefaults 设置默认配置值
func (m *Manager) setDefaults() {
	// 应用默认配置
	m.viper.SetDefault("app.name", "DirAI-Scan")
	m.viper.SetDefault("app.version", "1.0.0")
	m.viper.SetDefault("app.debug", false)

	// 日志默认配置
	m.viper.SetDefault("logging.level", "info")
	m.viper.SetDefault("logging.dir", "./logs")
	m.viper.SetDefault("logging.max_size", 100)
	m.viper.SetDefault("logging.max_backups", 3)
	m.viper.SetDefault("logging.max_age", 28)
	m.viper.SetDefault("logging.compress", true)

	// 扫描默认配置
	m.viper.SetDefault("scan.concurrency", 50)
	m.viper.SetDefault("scan.timeout", "5s")
	m.viper.SetDefault("scan.retries", 3)
	m.viper.SetDefault("scan.delay", "0ms")

	// AI默认配置
	m.viper.SetDefault("ai.enabled", false)
	m.viper.SetDefault("ai.endpoint", "http://localhost:11434")
	m.viper.SetDefault("ai.model", "deepseek-r1")
	m.viper.SetDefault("ai.timeout", "30s")
	m.viper.SetDefault("ai.max_retries", 3)

	// 输出默认配置
	m.viper.SetDefault("output.format", "json")
	m.viper.SetDefault("output.directory", "./results")
	m.viper.SetDefault("output.filename_template", "scan_results_{{.target}}_{{.timestamp}}")
}

// validate 验证配置
func (m *Manager) validate() error {
	// 验证日志级别
	validLogLevels := []string{"debug", "info", "warn", "error", "fatal"}
	if !contains(validLogLevels, m.config.Logging.Level) {
		return fmt.Errorf("invalid log level: %s", m.config.Logging.Level)
	}

	// 验证扫描超时
	if _, err := time.ParseDuration(m.config.Scan.Timeout); err != nil {
		return fmt.Errorf("invalid scan timeout: %s", m.config.Scan.Timeout)
	}

	// 验证AI超时
	if m.config.AI.Enabled {
		if _, err := time.ParseDuration(m.config.AI.Timeout); err != nil {
			return fmt.Errorf("invalid AI timeout: %s", m.config.AI.Timeout)
		}
	}

	// 验证输出目录
	if m.config.Output.Directory != "" {
		if err := os.MkdirAll(m.config.Output.Directory, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// 验证日志目录
	if m.config.Logging.Dir != "" {
		if err := os.MkdirAll(m.config.Logging.Dir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	return nil
}

// Get 获取配置
func (m *Manager) Get() *Config {
	return m.config
}

// GetString 获取字符串配置值
func (m *Manager) GetString(key string) string {
	return m.viper.GetString(key)
}

// GetInt 获取整数配置值
func (m *Manager) GetInt(key string) int {
	return m.viper.GetInt(key)
}

// GetBool 获取布尔配置值
func (m *Manager) GetBool(key string) bool {
	return m.viper.GetBool(key)
}

// GetDuration 获取时间间隔配置值
func (m *Manager) GetDuration(key string) time.Duration {
	return m.viper.GetDuration(key)
}

// Set 设置配置值
func (m *Manager) Set(key string, value interface{}) {
	m.viper.Set(key, value)
}

// Save 保存配置到文件
func (m *Manager) Save(configPath string) error {
	if configPath == "" {
		configPath = "./configs/dirmap.yaml"
	}

	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return m.viper.WriteConfigAs(configPath)
}

// Watch 监听配置文件变化
func (m *Manager) Watch(callback func()) {
	m.viper.WatchConfig()
	m.viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("Config file changed: %s\n", e.Name)
		if callback != nil {
			callback()
		}
	})
}

// contains 检查切片是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetScanTimeout 获取扫描超时时间
func (c *Config) GetScanTimeout() time.Duration {
	duration, _ := time.ParseDuration(c.Scan.Timeout)
	return duration
}

// GetScanDelay 获取扫描延迟时间
func (c *Config) GetScanDelay() time.Duration {
	duration, _ := time.ParseDuration(c.Scan.Delay)
	return duration
}

// GetAITimeout 获取AI超时时间
func (c *Config) GetAITimeout() time.Duration {
	duration, _ := time.ParseDuration(c.AI.Timeout)
	return duration
}

// GetDNSTimeout 获取DNS超时时间
func (c *Config) GetDNSTimeout() time.Duration {
	duration, _ := time.ParseDuration(c.Network.DNS.Timeout)
	return duration
}

// IsStatusCodeIncluded 检查状态码是否被包含
func (c *Config) IsStatusCodeIncluded(code int) bool {
	// 检查是否在排除列表中
	for _, excludeCode := range c.Scan.StatusCodes.Exclude {
		if code == excludeCode {
			return false
		}
	}

	// 如果包含列表为空，则包含所有状态码（除了排除的）
	if len(c.Scan.StatusCodes.Include) == 0 {
		return true
	}

	// 检查是否在包含列表中
	for _, includeCode := range c.Scan.StatusCodes.Include {
		if code == includeCode {
			return true
		}
	}

	return false
}

// IsResponseSizeValid 检查响应大小是否有效
func (c *Config) IsResponseSizeValid(size int64) bool {
	// 检查是否在排除大小列表中
	for _, excludeSize := range c.Scan.ResponseSize.ExcludeSizes {
		if size == excludeSize {
			return false
		}
	}

	// 检查大小范围
	if c.Scan.ResponseSize.Min > 0 && size < c.Scan.ResponseSize.Min {
		return false
	}

	if c.Scan.ResponseSize.Max > 0 && size > c.Scan.ResponseSize.Max {
		return false
	}

	return true
}

// 兼容性函数，保持向后兼容
var globalManager *Manager

// LoadConfig 加载配置文件（兼容性函数）
func LoadConfig(configPath string) error {
	if globalManager == nil {
		globalManager = NewManager()
	}
	return globalManager.Load(configPath)
}

// GetConfig 获取全局配置（兼容性函数）
func GetConfig() *Config {
	if globalManager == nil {
		globalManager = NewManager()
		// 尝试加载默认配置
		_ = globalManager.Load("")
	}
	return globalManager.Get()
}
