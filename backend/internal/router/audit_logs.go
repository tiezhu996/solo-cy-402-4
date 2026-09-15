package router

import (
	"cylawcase/internal/constants"
	"cylawcase/internal/middleware"

	"github.com/gin-gonic/gin"
)

// registerAuditLogRoutes 审计日志路由（仅管理员）。
func (r *Router) registerAuditLogRoutes(g *gin.RouterGroup) {
	logs := g.Group("/audit-logs")
	logs.Use(middleware.AuthRequired(r.cfg), middleware.RequireRole(constants.RoleAdmin))
	logs.GET("", r.auditLog.List)
}
