package util

import (
	"log/slog"
	"os"
	"time"
)

// NewLogger 创建 slog 结构化日志器。
func NewLogger(level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}

// DurationHours 将小时数转换为 time.Duration。
func DurationHours(hours int) time.Duration {
	return time.Duration(hours) * time.Hour
}
