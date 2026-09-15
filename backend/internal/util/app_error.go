package util

import "fmt"

// AppError 业务错误，携带统一错误码。
type AppError struct {
	Code    int
	Message string
}

// Error 实现 error 接口。
func (e *AppError) Error() string {
	return fmt.Sprintf("code=%d message=%s", e.Code, e.Message)
}

// NewAppError 构造业务错误。
func NewAppError(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Wrap 包装错误并附带上下文。
func Wrap(err error, format string, args ...any) error {
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}
