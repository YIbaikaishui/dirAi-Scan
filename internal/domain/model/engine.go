package model

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"

	"github.com/diraiscan/internal/pkg/config"
	"github.com/diraiscan/internal/pkg/utils"
)

// Result 扫描结果
type Result struct {
	URL           string            `json:"url"`
	StatusCode    int               `json:"status_code"`
	ContentLength int64             `json:"content_length"`
	Headers       map[string]string `json:"headers"`
	Body          string            `json:"body,omitempty"`
	Error         string            `json:"error,omitempty"`
	Time          time.Time         `json:"time"`
	ResponseTime  time.Duration     `json:"response_time"`
	TechStack     []string          `json:"tech_stack,omitempty"`
	BodyHash      string            `json:"body_hash,omitempty"` // 响应体hash
}

// Engine 扫描引擎
type Engine struct {
	config  *config.Config
	client  *fasthttp.Client
	results chan Result
	errors  chan error
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	baseURL string
}

// EngineOptions 引擎配置选项
type EngineOptions struct {
	Threads    int
	Timeout    time.Duration
	UserAgent  string
	Proxy      string
	RetryTimes int
	RetryDelay time.Duration
}

// NewEngine 创建新的扫描引擎
func NewEngine(cfg *config.Config) *Engine {
	ctx, cancel := context.WithCancel(context.Background())

	// 解析超时时间
	timeout, err := time.ParseDuration(cfg.Scan.Timeout)
	if err != nil {
		timeout = 5 * time.Second // 默认5秒
	}

	client := &fasthttp.Client{
		MaxConnsPerHost: cfg.Scan.Concurrency,
		ReadTimeout:     timeout,
		WriteTimeout:    timeout,
	}

	return &Engine{
		config:  cfg,
		client:  client,
		results: make(chan Result, cfg.Scan.Concurrency),
		errors:  make(chan error, cfg.Scan.Concurrency),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// NewEngine 使用选项创建新的扫描引擎
func NewEngineWithOptions(opts EngineOptions) *Engine {
	ctx, cancel := context.WithCancel(context.Background())

	client := &fasthttp.Client{
		MaxConnsPerHost: opts.Threads,
		ReadTimeout:     opts.Timeout,
		WriteTimeout:    opts.Timeout,
	}

	cfg := &config.Config{}
	cfg.Scan.Concurrency = opts.Threads
	cfg.Scan.Timeout = fmt.Sprintf("%ds", int(opts.Timeout.Seconds()))
	cfg.Scan.Retries = opts.RetryTimes
	cfg.Scan.Delay = fmt.Sprintf("%ds", int(opts.RetryDelay.Seconds()))

	return &Engine{
		config:  cfg,
		client:  client,
		results: make(chan Result, opts.Threads),
		errors:  make(chan error, opts.Threads),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Scan 扫描目标并返回结果
func (e *Engine) Scan(targets []string) ([]ScanResult, error) {
	if len(targets) == 0 {
		return nil, fmt.Errorf("没有指定扫描目标")
	}

	// 启动扫描
	err := e.Start(targets[0], targets[1:])
	if err != nil {
		return nil, err
	}

	// 收集结果
	var results []ScanResult
	for result := range e.results {
		// 将Result转换为ScanResult
		scanResult := ScanResult{
			URL:           result.URL,
			StatusCode:    result.StatusCode,
			ContentLength: result.ContentLength,
			Headers:       result.Headers,
			Body:          result.Body,
			Error:         result.Error,
			Time:          result.Time,
			ResponseTime:  result.ResponseTime,
			TechStack:     result.TechStack,
			BodyHash:      result.BodyHash,
		}
		results = append(results, scanResult)
	}

	return results, nil
}

// Start 开始扫描
func (e *Engine) Start(targetURL string, paths []string) error {
	_, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("无效的目标URL: %v", err)
	}

	// 保存基础URL
	e.baseURL = targetURL

	// 导入color包
	infoColor := color.New(color.FgHiBlue)
	valueColor := color.New(color.FgHiGreen, color.Bold)
	headerColor := color.New(color.FgHiCyan, color.Underline)

	// 打印扫描开始信息和目标
	fmt.Println()

	infoColor.Print("[*] 开始扫描 ")
	valueColor.Printf("%s\n", targetURL)

	infoColor.Print("[*] 共加载 ")
	valueColor.Printf("%d\n", len(paths))

	infoColor.Print("[*] 线程数: ")
	valueColor.Printf("%d\n", e.config.Scan.Concurrency)

	infoColor.Print("[*] 超时: ")
	valueColor.Printf("%s\n\n", e.config.Scan.Timeout)

	// 打印表头
	headerColor.Printf("%-4s %-7s %-40s %-15s %-10s\n", "状态", "码", "路径", "大小", "时间")
	fmt.Printf("%s\n", strings.Repeat("-", 80))

	utils.Info("开始扫描",
		zap.String("target", targetURL),
		zap.Int("threads", e.config.Scan.Concurrency),
		zap.Int("paths", len(paths)),
	)

	// 创建工作池
	pathChan := make(chan string, e.config.Scan.Concurrency)

	// 启动工作协程
	for i := 0; i < e.config.Scan.Concurrency; i++ {
		e.wg.Add(1)
		go e.worker(pathChan)
	}

	// 发送任务
	go func() {
		for _, path := range paths {
			select {
			case pathChan <- path:
			case <-e.ctx.Done():
				return
			}
		}
		close(pathChan)
	}()

	// 等待所有工作完成
	go func() {
		e.wg.Wait()
		close(e.results)
		close(e.errors)
	}()

	return nil
}

// worker 工作协程
func (e *Engine) worker(paths <-chan string) {
	defer e.wg.Done()

	for path := range paths {
		select {
		case <-e.ctx.Done():
			return
		default:
			e.scanPath(path)
		}
	}
}

// scanPath 扫描单个路径
func (e *Engine) scanPath(path string) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	// 构建完整URL
	targetURL := path
	if !strings.HasPrefix(path, "http") {
		// 获取基础URL
		baseURL, _ := url.Parse(e.baseURL)
		if baseURL != nil {
			targetURL = baseURL.String()
			if !strings.HasSuffix(targetURL, "/") && !strings.HasPrefix(path, "/") {
				targetURL += "/"
			}
			targetURL += strings.TrimPrefix(path, "/")
		}
	}

	// 设置请求
	req.SetRequestURI(targetURL)
	req.Header.SetMethod("GET")
	// 使用默认User-Agent
	req.Header.SetUserAgent("dirAi-Scan/1.0")

	startTime := time.Now()
	// 发送请求
	err := e.client.Do(req, resp)
	duration := time.Since(startTime)

	if err != nil {
		e.errors <- fmt.Errorf("请求失败 %s: %v", path, err)
		return
	}

	// 处理响应
	body := resp.Body()
	result := Result{
		URL:           path,
		StatusCode:    resp.StatusCode(),
		ContentLength: int64(len(body)),
		ResponseTime:  duration,
		Headers:       make(map[string]string),
		BodyHash:      utils.HashMD5(body),
	}

	// 收集响应头
	resp.Header.VisitAll(func(key, value []byte) {
		result.Headers[string(key)] = string(value)
	})

	// 保存响应体（可根据需要配置）
	// result.Body = string(body)

	// 发送结果到通道
	e.results <- result
}

// GetResults 获取扫描结果通道
func (e *Engine) GetResults() <-chan Result {
	return e.results
}

// GetErrors 获取错误通道
func (e *Engine) GetErrors() <-chan error {
	return e.errors
}

// Stop 停止扫描
func (e *Engine) Stop() {
	e.cancel()
	e.wg.Wait()
}

// Wait 等待扫描完成
func (e *Engine) Wait() {
	e.wg.Wait()
	close(e.results)
	close(e.errors)
}
