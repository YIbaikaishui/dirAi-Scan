package sdk

import (
	"time"

	"github.com/diraiscan/internal/domain/model"
)

type ScanOptions struct {
	Threads    int
	Timeout    time.Duration
	UserAgent  string
	Proxy      string
	RetryTimes int
	RetryDelay time.Duration
}

func DefaultScanOptions() *ScanOptions {
	return &ScanOptions{
		Threads:    100,
		Timeout:    30 * time.Second,
		UserAgent:  "", // User-Agent will be loaded from config or data/user-agents.txt
		RetryTimes: 3,
		RetryDelay: 5 * time.Second,
	}
}

type Scanner interface {
	Scan(targets []string) ([]model.Result, error)
	SetOptions(opts *ScanOptions)
	GetEngine() *model.Scanner
}

func NewScanner() Scanner {
	return &scannerImpl{
		opts: DefaultScanOptions(),
	}
}

type scannerImpl struct {
	opts   *ScanOptions
	engine *model.Scanner
}

func (s *scannerImpl) Scan(targets []string) ([]model.Result, error) {
	// 创建引擎（如果尚未创建）
	if s.engine == nil {
		s.engine = model.NewScannerWithOptions(model.ScannerOptions{
			Threads:    s.opts.Threads,
			Timeout:    s.opts.Timeout,
			UserAgent:  s.opts.UserAgent,
			Proxy:      s.opts.Proxy,
			RetryTimes: s.opts.RetryTimes,
			RetryDelay: s.opts.RetryDelay,
		})
	}

	return s.engine.Scan(targets)
}

func (s *scannerImpl) SetOptions(opts *ScanOptions) {
	s.opts = opts
}

// GetEngine 返回底层的扫描引擎
func (s *scannerImpl) GetEngine() *model.Scanner {
	// 如果引擎尚未创建，则创建一个
	if s.engine == nil {
		s.engine = model.NewScannerWithOptions(model.ScannerOptions{
			Threads:    s.opts.Threads,
			Timeout:    s.opts.Timeout,
			UserAgent:  s.opts.UserAgent,
			Proxy:      s.opts.Proxy,
			RetryTimes: s.opts.RetryTimes,
			RetryDelay: s.opts.RetryDelay,
		})
	}

	return s.engine
}
