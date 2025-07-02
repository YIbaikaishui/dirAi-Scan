package sdk

import (
	"github.com/diraiscan/internal/infra/ai"
	"github.com/diraiscan/internal/domain/model"
)

type AIOptions struct {
	Enabled  bool
	Model    string
	APIKey   string
	Endpoint string
	Local    bool
}

func DefaultAIOptions() *AIOptions {
	return &AIOptions{
		Enabled:  false,
		Model:    "qwen3:4b",
		Endpoint: "http://localhost:11434",
		Local:    true,
	}
}

type Analyzer interface {
	Analyze(results []model.Result) ([]ai.ScanResultAnalysis, error)
	AnalyzeSingle(result model.Result) ([]ai.ScanResultAnalysis, error)
	SetOptions(opts *AIOptions)
}

func NewAnalyzer() Analyzer {
	return &analyzerImpl{
		opts: DefaultAIOptions(),
	}
}

type analyzerImpl struct {
	opts *AIOptions
}

func (a *analyzerImpl) Analyze(results []model.Result) ([]ai.ScanResultAnalysis, error) {
	if !a.opts.Enabled {
		return nil, nil
	}

	analyzer := ai.NewAIAnalyzer(a.opts.Endpoint, a.opts.Model)

	return analyzer.AnalyzeResults(results)
}

func (a *analyzerImpl) AnalyzeSingle(result model.Result) ([]ai.ScanResultAnalysis, error) {
	if !a.opts.Enabled {
		return nil, nil
	}

	analyzer := ai.NewAIAnalyzer(a.opts.Endpoint, a.opts.Model)

	// 将单个结果转换为切片并分析
	results := []model.Result{result}
	return analyzer.AnalyzeResults(results)
}

func (a *analyzerImpl) SetOptions(opts *AIOptions) {
	a.opts = opts
}
