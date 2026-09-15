package handler

import (
	"log/slog"
	"net/http"

	"cylawcase/internal/dto"
	"cylawcase/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuditLogHandler 审计日志 HTTP 处理器。
type AuditLogHandler struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewAuditLogHandler 构造审计日志处理器。
func NewAuditLogHandler(db *gorm.DB, logger *slog.Logger) *AuditLogHandler {
	return &AuditLogHandler{db: db, logger: logger}
}

// List 审计日志分页查询。
func (h *AuditLogHandler) List(c *gin.Context) {
	var q dto.PageQuery
	_ = c.ShouldBindQuery(&q)
	q.Normalize()
	var list []model.AuditLog
	var total int64
	if err := h.db.Model(&model.AuditLog{}).Count(&total).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 50000, "服务器内部错误")
		return
	}
	if err := h.db.Order("id DESC").Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Find(&list).Error; err != nil {
		Fail(c, http.StatusInternalServerError, 50000, "服务器内部错误")
		return
	}
	OK(c, pageResponse(list, total, q.Page, q.PageSize))
}
