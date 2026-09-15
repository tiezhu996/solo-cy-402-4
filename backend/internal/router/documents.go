package router

import (
	"cylawcase/internal/middleware"

	"github.com/gin-gonic/gin"
)

// registerDocumentRoutes 文档路由。
func (r *Router) registerDocumentRoutes(g *gin.RouterGroup) {
	docs := g.Group("/documents")
	docs.Use(middleware.AuthRequired(r.cfg))
	docs.GET("", r.document.List)
	docs.POST("", r.document.Create)
	docs.GET("/by-case/:id", r.document.ListByCase)
	docs.DELETE("/:id", r.document.Delete)
}
