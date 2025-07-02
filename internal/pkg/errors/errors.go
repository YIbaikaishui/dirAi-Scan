package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// ErrorCode 错误代码类型
type ErrorCode string

const (
	// 通用错误
	ErrCodeInternal    ErrorCode = "INTERNAL_ERROR"
	ErrCodeInvalidArgs ErrorCode = "INVALID_ARGS"
	ErrCodeNotFound    ErrorCode = "NOT_FOUND"
	ErrCodeTimeout     ErrorCode = "TIMEOUT"

	// 配置错误
	ErrCodeConfigLoad   ErrorCode = "CONFIG_LOAD_ERROR"
	ErrCodeConfigParse  ErrorCode = "CONFIG_PARSE_ERROR"
	ErrCodeConfigInvalid ErrorCode = "CONFIG_INVALID"

	// 扫描错误
	ErrCodeScanFailed    ErrorCode = "SCAN_FAILED"
	ErrCodeScanTimeout   ErrorCode = "SCAN_TIMEOUT"
	ErrCodeScanCancelled ErrorCode = "SCAN_CANCELLED"

	// AI错误
	ErrCodeAIUnavailable ErrorCode = "AI_UNAVAILABLE"
	ErrCodeAITimeout     ErrorCode = "AI_TIMEOUT"
	ErrCodeAIFailed      ErrorCode = "AI_FAILED"

	// 网络错误
	ErrCodeNetworkError ErrorCode = "NETWORK_ERROR"
	ErrCodeDNSError     ErrorCode = "DNS_ERROR"
	ErrCodeConnRefused  ErrorCode = "CONNECTION_REFUSED"
)

// AppError 应用错误
type AppError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Cause   error     `json:"cause,omitempty"`
	Stack   string    `json:"stack,omitempty"`
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap 实现errors.Unwrap接口
func (e *AppError) Unwrap() error {
	return e.Cause
}

// New 创建新的应用错误
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Stack:   getStack(),
	}
}

// Wrap 包装错误
func Wrap(err error, code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   err,
		Stack:   getStack(),
	}
}

// Wrapf 格式化包装错误
func Wrapf(err error, code ErrorCode, format string, args ...interface{}) *AppError {
	return &AppError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Cause:   err,
		Stack:   getStack(),
	}
}

// IsCode 检查错误是否为指定代码
func IsCode(err error, code ErrorCode) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == code
	}
	return false
}

// GetCode 获取错误代码
func GetCode(err error) ErrorCode {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}
	return ErrCodeInternal
}

// getStack 获取调用栈
func getStack() string {
	var stack []string
	for i := 2; i < 10; i++ { // 跳过getStack和调用者
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// 只保留文件名，不要完整路径
		parts := strings.Split(file, "/")
		filename := parts[len(parts)-1]
		stack = append(stack, fmt.Sprintf("%s:%d", filename, line))
	}
	return strings.Join(stack, " -> ")
}

// 预定义的常用错误
var (
	ErrInvalidArgs    = New(ErrCodeInvalidArgs, "invalid arguments")
	ErrNotFound       = New(ErrCodeNotFound, "resource not found")
	ErrTimeout        = New(ErrCodeTimeout, "operation timeout")
	ErrConfigLoad     = New(ErrCodeConfigLoad, "failed to load configuration")
	ErrScanFailed     = New(ErrCodeScanFailed, "scan operation failed")
	ErrAIUnavailable  = New(ErrCodeAIUnavailable, "AI service unavailable")
	ErrNetworkError   = New(ErrCodeNetworkError, "network error")
)