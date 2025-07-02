package model

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/diraiscan/internal/pkg/utils"

	"github.com/valyala/fasthttp"
)

type Scanner struct {
	Target            string
	Wordlist          []string
	Concurrency       int
	Timeout           time.Duration
	Client            *fasthttp.Client
	RateLimit         *RateLimiter
	Results           []Result
	mu                sync.Mutex
	StatusCodes       map[int]bool
	ExcludeStatus     map[int]bool
	ExcludeSize       int
	Retries           int
	Delay             time.Duration
	AIEnabled         bool
	AIModel           string
	AIEndpoint        string
	Mode              string
	TechDetect        string
	InvalidSignatures []struct {
		StatusCode    int
		ContentLength int
		BodyHash      string
	}
	UserAgents []string
}

type RateLimiter struct {
	Rate   int
	Per    time.Duration
	tokens chan struct{}
	stop   chan struct{}
}

func NewRateLimiter(rate int, per time.Duration) *RateLimiter {
	rl := &RateLimiter{
		Rate:   rate,
		Per:    per,
		tokens: make(chan struct{}, rate),
		stop:   make(chan struct{}),
	}

	for i := 0; i < rate; i++ {
		rl.tokens <- struct{}{}
	}

	go func() {
		ticker := time.NewTicker(per / time.Duration(rate))
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				select {
				case rl.tokens <- struct{}{}:
				default:
				}

			case <-rl.stop:
				return
			}

		}

	}()

	return rl
}

func (rl *RateLimiter) Take() {
	<-rl.tokens
}

func (rl *RateLimiter) Stop() {
	close(rl.stop)
}

func NewScanner(target string, options ...func(*Scanner)) *Scanner {
	_, err := url.Parse(target)
	if err != nil {
		panic(fmt.Sprintf("无效的目标URL: %s", err))
	}

	s := &Scanner{
		Target:      target,
		Concurrency: 50,
		Timeout:     5 * time.Second,
		Client: &fasthttp.Client{
			MaxConnsPerHost: 1000,
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    5 * time.Second,
		},
		StatusCodes:   map[int]bool{200: true, 403: true, 500: true},
		ExcludeStatus: map[int]bool{404: true},
		Retries:       3,
		Mode:          "fast",
		TechDetect:    "quick",
		UserAgents:    loadUserAgents(utils.AbsPath("data/user-agents.txt")),
	}

	for _, option := range options {
		option(s)
	}
	return s
}

func loadUserAgents(path string) []string {
	lines, err := utils.ReadLines(path)
	if err != nil {
		// Fallback to a default user agent if the file cannot be read
		return []string{"Mozilla/5.0 (compatible; DirAiScan/1.0)"}
	}
	return lines
}

func WithConcurrency(n int) func(*Scanner) {
	return func(s *Scanner) {
		s.Concurrency = n
	}
}

func WithTimeout(d time.Duration) func(*Scanner) {
	return func(s *Scanner) {
		s.Timeout = d
		s.Client.ReadTimeout = d
		s.Client.WriteTimeout = d
	}
}

func WithWordlist(wordlist []string) func(*Scanner) {
	return func(s *Scanner) {
		s.Wordlist = wordlist
	}
}

func WithRateLimit(rate int, per time.Duration) func(*Scanner) {
	return func(s *Scanner) {
		s.RateLimit = NewRateLimiter(rate, per)
	}
}

func WithStatusCodes(codes []int) func(*Scanner) {
	return func(s *Scanner) {
		s.StatusCodes = make(map[int]bool)
		for _, code := range codes {
			s.StatusCodes[code] = true
		}
	}
}

func WithExcludeStatus(codes []int) func(*Scanner) {
	return func(s *Scanner) {
		s.ExcludeStatus = make(map[int]bool)
		for _, code := range codes {
			s.ExcludeStatus[code] = true
		}
	}
}

func WithAI(enabled bool, model string, endpoint string) func(*Scanner) {
	return func(s *Scanner) {
		s.AIEnabled = enabled
		s.AIModel = model
		s.AIEndpoint = endpoint
	}
}

func WithExcludeSize(size int) func(*Scanner) {
	return func(s *Scanner) {
		s.ExcludeSize = size
	}
}

func WithRetries(retries int) func(*Scanner) {
	return func(s *Scanner) {
		s.Retries = retries
	}
}

func WithDelay(delay time.Duration) func(*Scanner) {
	return func(s *Scanner) {
		s.Delay = delay
	}
}

func WithMode(mode string) func(*Scanner) {
	return func(s *Scanner) {
		s.Mode = mode
	}
}

func WithTechDetect(techDetect string) func(*Scanner) {
	return func(s *Scanner) {
		s.TechDetect = techDetect
	}
}

func WithUserAgents(userAgents []string) func(*Scanner) {
	return func(s *Scanner) {
		s.UserAgents = userAgents
	}
}

func WithInvalidSignatures(signatures []struct {
	StatusCode    int
	ContentLength int
	BodyHash      string
}) func(*Scanner) {
	return func(s *Scanner) {
		s.InvalidSignatures = signatures
	}
}

// 定义回调函数类型，用于实时处理扫描结果
type ResultCallback func(result Result)

// Start 开始扫描并返回所有结果
func (s *Scanner) Start() []Result {
	// 使用默认的回调函数（不做任何操作）
	return s.StartWithCallback(nil)
}

// StartWithCallback 开始扫描，并通过回调函数实时处理结果
func (s *Scanner) StartWithCallback(callback ResultCallback) []Result {
	var wg sync.WaitGroup
	resultsChan := make(chan Result, s.Concurrency)

	// 创建一个上下文，用于控制所有goroutine的生命周期
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // 确保在函数退出时取消所有goroutine

	// 计数器，用于显示进度
	totalPaths := len(s.Wordlist)
	processedPaths := 0
	var processMutex sync.Mutex

	// 启动工作协程池
	for i := 0; i < s.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 从 nextPath 获取路径，该函数现在接受上下文
			for path := range s.nextPath(ctx) { // 直接在循环中处理来自 nextPath 的路径
				select {
				case <-ctx.Done(): // 检查上下文是否已取消
					return
				default:
					result := s.scanPath(path)

					// 更新进度计数
					processMutex.Lock()
					processedPaths++
					processMutex.Unlock()

					if result != nil {
						select {
						case resultsChan <- *result:
							// 如果提供了回调函数，则调用它
							if callback != nil {
								callback(*result)
							}
						case <-ctx.Done(): // 再次检查，以防在发送结果时取消
							return
						}
					}
				}
			}
		}()
	}

	// 启动进度显示协程
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				processMutex.Lock()
				current := processedPaths
				processMutex.Unlock()

				if current >= totalPaths {
					return
				}

				// 计算并显示进度百分比
				percent := float64(current) / float64(totalPaths) * 100
				fmt.Printf("\r[进度] %.2f%% (%d/%d)    ", percent, current, totalPaths)
			case <-ctx.Done():
				return
			}
		}
	}()

	// 单独的 goroutine 来等待所有工作协程完成并关闭 resultsChan
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// 收集结果 (这个 goroutine 保持不变，但现在它会在 resultsChan 关闭后正确退出)
	for result := range resultsChan {
		s.mu.Lock()
		s.Results = append(s.Results, result)
		s.mu.Unlock()
	}

	// 清除进度显示的最后一行
	fmt.Print("\r                                        \r")

	// 停止速率限制器（如果存在）
	if s.RateLimit != nil {
		s.RateLimit.Stop()
	}
	return s.Results
}

func (s *Scanner) nextPath(ctx context.Context) chan string {
	pathChan := make(chan string, s.Concurrency) // 使用带缓冲的通道
	go func() {
		defer close(pathChan)
		for _, path := range s.Wordlist {
			select {
			case pathChan <- path:
			case <-ctx.Done(): // 检查上下文是否已取消
				return
			}
		}
	}()
	return pathChan
}

func (s *Scanner) scanPath(path string) *Result {
	fullURL := fmt.Sprintf("%s%s", s.Target, path)

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(fullURL)
	req.Header.SetMethod("GET")

	// 随机User-Agent
	// User-Agent will be loaded from data/user-agents.txt
	// Ensure userAgents is loaded in the NewEngine or relevant initialization part
	if len(s.UserAgents) > 0 {
		// rand.Seed(time.Now().UnixNano()) // rand.Seed 应该在程序启动时调用一次
		req.Header.SetUserAgent(s.UserAgents[rand.Intn(len(s.UserAgents))])
	} else {
		// Fallback or error handling if userAgents are not loaded
		req.Header.SetUserAgent("Mozilla/5.0 (compatible; DirAiScan/1.0)")
	}

	// 速率限制
	if s.RateLimit != nil {
		s.RateLimit.Take()
	}

	start := time.Now()
	err := s.Client.DoTimeout(req, resp, s.Timeout)
	responseTime := time.Since(start)

	if err != nil {
		// 处理错误，例如超时、连接问题等
		// fmt.Printf("Error scanning %s: %v\n", fullURL, err)
		return nil
	}

	statusCode := resp.StatusCode()
	contentLength := len(resp.Body())

	// 检查排除状态码
	if s.ExcludeStatus[statusCode] {
		return nil
	}

	// 检查排除大小
	if s.ExcludeSize > 0 && contentLength == s.ExcludeSize {
		return nil
	}

	// 检查无效签名
	bodyHash := utils.HashMD5(resp.Body())
	for _, sig := range s.InvalidSignatures {
		if sig.StatusCode == statusCode && sig.ContentLength == contentLength && sig.BodyHash == bodyHash {
			return nil
		}
	}

	// 检查是否是期望的状态码
	if len(s.StatusCodes) > 0 && !s.StatusCodes[statusCode] {
		return nil
	}

	headers := make(map[string]string)
	resp.Header.VisitAll(func(key, value []byte) {
		headers[string(key)] = string(value)
	})

	// 技术栈识别
	var techStack []string
	if s.TechDetect != "none" {
		techStack = DetectTech(resp.Body(), headers, s.TechDetect)
	}

	return &Result{
		URL:           fullURL,
		StatusCode:    statusCode,
		ContentLength: int64(contentLength),
		ResponseTime:  responseTime,
		Headers:       headers,
		TechStack:     techStack,
		BodyHash:      bodyHash,
	}
}

// DetectTech 识别技术栈
func DetectTech(body []byte, headers map[string]string, mode string) []string {
	var detected []string

	// 示例：通过响应头识别技术
	if server, ok := headers["Server"]; ok {
		if strings.Contains(server, "nginx") {
			detected = append(detected, "Nginx")
		} else if strings.Contains(server, "Apache") {
			detected = append(detected, "Apache")
		}
	}

	// 示例：通过响应体识别技术
	if strings.Contains(string(body), "WordPress") {
		detected = append(detected, "WordPress")
	}

	// 根据模式进行更深入的检测
	if mode == "full" {
		// 这里可以添加更复杂的正则表达式匹配、指纹识别等
		// 例如，检查特定的JS文件、HTML注释、Meta标签等
		if strings.Contains(string(body), "React") {
			detected = append(detected, "React")
		}

		if strings.Contains(string(body), "Vue") {
			detected = append(detected, "Vue.js")
		}
	}

	return detected
}
