package service

import (
	"context"
	"github.com/diraiscan/internal/domain/model"
)

// AIService AI服务接口
type AIService interface {
	// AnalyzeResults 分析扫描结果
	AnalyzeResults(ctx context.Context, results []model.Result) ([]AIAnalysis, error)
	// AnalyzeResult 分析单个结果
	AnalyzeResult(ctx context.Context, result model.Result) (*AIAnalysis, error)
	// IsAvailable 检查AI服务是否可用
	IsAvailable(ctx context.Context) bool
}

// AIAnalysis AI分析结果
type AIAnalysis struct {
	OriginalResult model.Result
	RiskLevel      string   // low, medium, high
	Findings       []string
	Suggestions    []string
	Confidence     float64
}

// DictionaryService 字典服务接口
type DictionaryService interface {
	// LoadDictionary 加载字典
	LoadDictionary(path string) ([]string, error)
	// GetPaths 获取扫描路径
	GetPaths() []string
	// GetUserAgents 获取用户代理列表
	GetUserAgents() []string
	// GetBlacklist 获取黑名单
	GetBlacklist(statusCode int) []string
}

// SchedulerService 调度服务接口
type SchedulerService interface {
	// Schedule 调度任务
	Schedule(ctx context.Context, task Task) error
	// GetStatus 获取任务状态
	GetStatus(taskID string) (*TaskStatus, error)
	// Cancel 取消任务
	Cancel(taskID string) error
}

// Task 任务定义
type Task struct {
	ID       string
	Type     string
	Target   string
	Options  map[string]interface{}
	Priority int
}

// TaskStatus 任务状态
type TaskStatus struct {
	ID       string
	Status   string // pending, running, completed, failed
	Progress float64
	Error    error
}