package dictionary

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Dictionary 表示扫描字典
type Dictionary struct {
	Paths      []string
	Priorities map[string]int
	mu         sync.RWMutex
}

// NewDictionary 创建新的字典实例
func NewDictionary() *Dictionary {
	return &Dictionary{
		Paths:      make([]string, 0),
		Priorities: make(map[string]int),
	}
}

// LoadFromFile 从文件加载字典
func (d *Dictionary) LoadFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		path := strings.TrimSpace(scanner.Text())
		if path != "" && !strings.HasPrefix(path, "#") {
			d.AddPath(path)
		}
	}

	return scanner.Err()
}

// LoadBuiltinDictionary 加载内置字典
func (d *Dictionary) LoadBuiltinDictionary(dictType string) error {
	// 根据不同类型加载不同的内置字典
	var dictPath string
	switch dictType {
	case "common":
		dictPath = "dictionaries/common.txt"
	case "api":
		dictPath = "dictionaries/api.txt"
	case "admin":
		dictPath = "dictionaries/admin.txt"
	default:
		dictPath = "dictionaries/common.txt"
	}

	// 获取可执行文件所在目录
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	execDir := filepath.Dir(execPath)
	return d.LoadFromFile(filepath.Join(execDir, dictPath))
}

// AddPath 添加路径到字典
func (d *Dictionary) AddPath(path string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 检查是否已存在
	for _, p := range d.Paths {
		if p == path {
			return
		}
	}

	d.Paths = append(d.Paths, path)
	// 初始优先级为0
	d.Priorities[path] = 0
}

// GetPaths 获取所有路径（按优先级排序）
func (d *Dictionary) GetPaths() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// 复制路径列表
	paths := make([]string, len(d.Paths))
	copy(paths, d.Paths)

	// 按优先级排序
	sort.Slice(paths, func(i, j int) bool {
		return d.Priorities[paths[i]] > d.Priorities[paths[j]]
	})

	return paths
}

// UpdatePriority 更新路径优先级
func (d *Dictionary) UpdatePriority(path string, delta int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.Priorities[path]; exists {
		d.Priorities[path] += delta
	}
}

// GenerateAdaptivePaths 生成自适应路径
func (d *Dictionary) GenerateAdaptivePaths(patterns []string, extensions []string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 根据模式和扩展名生成新路径
	for _, pattern := range patterns {
		// 添加无扩展名版本
		d.Paths = append(d.Paths, pattern)
		d.Priorities[pattern] = 0

		// 添加带扩展名版本
		for _, ext := range extensions {
			path := pattern
			if !strings.HasSuffix(pattern, ".") && !strings.HasPrefix(ext, ".") {
				path += "."
			}
			path += ext

			d.Paths = append(d.Paths, path)
			d.Priorities[path] = 0
		}
	}
}

// AdjustPrioritiesByResponse 根据响应调整优先级
func (d *Dictionary) AdjustPrioritiesByResponse(path string, statusCode int, contentLength int) {
	// 根据响应特征调整优先级
	switch {
	case statusCode >= 200 && statusCode < 300:
		// 成功响应，提高相似路径优先级
		d.UpdatePriority(path, 10)
		d.boostSimilarPaths(path, 5)

	case statusCode == 403:
		// 禁止访问，可能是敏感目录
		d.UpdatePriority(path, 8)
		d.boostSimilarPaths(path, 4)

	case statusCode >= 500:
		// 服务器错误，可能是漏洞
		d.UpdatePriority(path, 6)
		d.boostSimilarPaths(path, 3)
	}
}

// boostSimilarPaths 提升相似路径的优先级
func (d *Dictionary) boostSimilarPaths(path string, boost int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 提取路径的前缀和后缀
	parts := strings.Split(path, "/")
	if len(parts) <= 1 {
		return
	}

	// 获取最后一部分和前缀
	lastPart := parts[len(parts)-1]
	prefix := strings.Join(parts[:len(parts)-1], "/")

	// 提升具有相同前缀的路径
	for p := range d.Priorities {
		if strings.HasPrefix(p, prefix) && p != path {
			d.Priorities[p] += boost / 2
		}
	}

	// 提升具有相似后缀的路径
	extParts := strings.Split(lastPart, ".")
	if len(extParts) > 1 {
		ext := extParts[len(extParts)-1]
		for p := range d.Priorities {
			if strings.HasSuffix(p, "."+ext) && p != path {
				d.Priorities[p] += boost / 2
			}
		}
	}
}

// GenerateMutations 生成变异路径
func (d *Dictionary) GenerateMutations(techStack []string) {
	// 根据检测到的技术栈生成特定变异
	var patterns []string
	var extensions []string

	// 根据技术栈调整生成策略
	for _, tech := range techStack {
		switch {
		case strings.Contains(strings.ToLower(tech), "php"):
			extensions = append(extensions, "php", "php.bak", "php~", "php.old")
			patterns = append(patterns, "config", "admin", "backup", "db")

		case strings.Contains(strings.ToLower(tech), "asp"):
			extensions = append(extensions, "asp", "aspx", "ashx", "asmx")
			patterns = append(patterns, "web", "admin", "backup", "db")

		case strings.Contains(strings.ToLower(tech), "java"):
			extensions = append(extensions, "jsp", "do", "action", "java")
			patterns = append(patterns, "servlet", "admin", "backup", "db")

		case strings.Contains(strings.ToLower(tech), "node") || strings.Contains(strings.ToLower(tech), "express"):
			extensions = append(extensions, "js", "json", "node")
			patterns = append(patterns, "api", "admin", "backup", "db")

		case strings.Contains(strings.ToLower(tech), "django") || strings.Contains(strings.ToLower(tech), "python"):
			extensions = append(extensions, "py", "pyc", "pyo")
			patterns = append(patterns, "admin", "static", "media", "db")
		}
	}

	// 去重
	extensions = uniqueStrings(extensions)
	patterns = uniqueStrings(patterns)

	// 生成变异路径
	d.GenerateAdaptivePaths(patterns, extensions)
}

// uniqueStrings 去除字符串切片中的重复项
func uniqueStrings(input []string) []string {
	unique := make(map[string]bool)
	for _, s := range input {
		unique[s] = true
	}

	result := make([]string, 0, len(unique))
	for s := range unique {
		result = append(result, s)
	}

	return result
}