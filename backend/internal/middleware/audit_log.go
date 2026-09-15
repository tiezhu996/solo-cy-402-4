package middleware

import (
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"cylawcase/internal/constants"
	"cylawcase/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuditLog 操作审计日志中间件，记录写操作。
func AuditLog(db *gorm.DB, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
		c.Next()
		if c.Writer.Status() >= 400 {
			return
		}
		path := c.FullPath()
		entityType := "unknown"
		seg := strings.Split(strings.TrimPrefix(path, "/api/v1/"), "/")
		if len(seg) > 0 && seg[0] != "" {
			entityType = seg[0]
		}
		detail := map[string]any{"method": c.Request.Method, "path": path}
		if b, ok := c.Get("audit_detail"); ok {
			detail["body"] = b
		}
		raw, _ := json.Marshal(detail)
		entry := &model.AuditLog{
			OperatorID: GetUserID(c), OperatorName: GetUsername(c),
			Action: c.Request.Method, EntityType: entityType,
			EntityID: c.Param("id"), Detail: string(raw), IP: c.ClientIP(), CreatedAt: time.Now(),
		}
		if err := db.Create(entry).Error; err != nil {
			logger.Error(constants.LogAuditWriteFailed, "error", err.Error())
		}
	}
}
