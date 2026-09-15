package router

import (
	"cylawcase/internal/middleware"

	"github.com/gin-gonic/gin"
)

// registerClientRoutes 客户路由。
func (r *Router) registerClientRoutes(g *gin.RouterGroup) {
	clients := g.Group("/clients")
	clients.Use(middleware.AuthRequired(r.cfg))
	clients.GET("", r.client.List)
	clients.GET("/:id", r.client.Get)
	clients.POST("", r.client.Create)
	clients.PUT("/:id", r.client.Update)
	clients.DELETE("/:id", r.client.Delete)
}
