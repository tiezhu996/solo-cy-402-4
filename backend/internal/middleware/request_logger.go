package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger 请求日志，包含 request_id、method、path、status、latency_ms。
func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		rid := c.Writer.Header().Get("X-Request-Id")
		if v, ok := c.Get("request_id"); ok {
			if s, ok := v.(string); ok && s != "" {
				rid = s
			}
		}
		logger.Info("http request", "request_id", rid,
			"method", c.Request.Method, "path", c.FullPath(),
			"status", c.Writer.Status(), "latency_ms", time.Since(start).Milliseconds())
	}
}
